package server

import (
	"errors"
	"github.com/axgle/mahonia"
	"io"
	"net"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

const (
	connStatusDispatching int32 = iota
	connStatusReading
	connStatusShutdown     // Closed by server.
	connStatusWaitShutdown // Notified by server to close.
)

type clientConn struct {
	bufReadConn  *bufferedReadConn
	server       *Server
	lastActive   time.Time
	status       int32
	connectionID uint64

	deviceIdentifierList sync.Map
}

func newClientConn(s *Server) (*clientConn, error) {

	return &clientConn{
		server:       s,
		status:       connStatusDispatching,
		connectionID: 11111111,
		lastActive:   time.Now(),
	}, nil
}

func (cc *clientConn) setConn(conn net.Conn) {
	cc.bufReadConn = newBufferedReadConn(conn)
}

func (cc *clientConn) Run() {
	defer func() {
		r := recover()
		if r != nil {
			logrus.Errorf("connection running loop panic, err: %+v", r)
		}

		if atomic.LoadInt32(&cc.status) != connStatusShutdown {
			err := cc.Close()
			if err != nil {
				logrus.Errorf("close connection err: %s", err)
			}
		}
	}()

	for {
		if !atomic.CompareAndSwapInt32(&cc.status, connStatusDispatching, connStatusReading) ||
			atomic.LoadInt32(&cc.status) == connStatusWaitShutdown {
			return
		}

		data, err := cc.readPacket()
		if err != nil {
			if !errors.Is(err, io.EOF) {
				logrus.Errorf("read packet failed, err: %s", err)

				if !atomic.CompareAndSwapInt32(&cc.status, connStatusReading, connStatusDispatching) {
					return
				}

				continue
			}

			return
		}

		if !atomic.CompareAndSwapInt32(&cc.status, connStatusReading, connStatusDispatching) {
			return
		}

		if err := cc.dispatch(data); err != nil {
			logrus.Errorf("dispatch data error: %s", err)
		}
	}
}

func (cc *clientConn) Close() error {
	logrus.Debugf("closing connection, remote address: %s, connection id: %d",
		cc.bufReadConn.RemoteAddr(),
		cc.connectionID)

	cc.deviceIdentifierList.Range(func(key, value interface{}) bool {
		if deviceIdentifier, ok := key.(string); ok {
			cc.server.rwlock.Lock()
			delete(cc.server.deviceConnectionID, deviceIdentifier)
			cc.server.rwlock.Unlock()
		}

		return true
	})

	cc.server.rwlock.Lock()
	delete(cc.server.clients, cc.connectionID)
	cc.server.rwlock.Unlock()

	err := cc.bufReadConn.Close()

	atomic.StoreInt32(&cc.status, connStatusShutdown)

	return err
}

func (cc *clientConn) ShutdownOrNotify() bool {
	if atomic.CompareAndSwapInt32(&cc.status, connStatusReading, connStatusShutdown) {
		return true
	}

	atomic.StoreInt32(&cc.status, connStatusWaitShutdown)

	return false
}

const (
	DataTypeLen          = 4                     // 数据类型长度
	HeaderLenIdentifyLen = 4                     // header长度标识长度
	TimeLen              = 19                    // 时间标识长度
	CheckLen             = 2                     // 校验码长度
	HeaderSplit          = "@@@"                 // header和data的分割符号
	DataSplit            = "tek"                 // data和校验码的分割符号
	EndTag               = "####"                // 结尾标识
	TimeFormat           = "2006-01-02 15:04:05" // 时间格式
)

type NeamMsg struct {
	Header string
	Data   string
}

func (cc *clientConn) readPacket() (*NeamMsg, error) {
	headDataByte, err := cc.read(HeaderSplit)
	if err != nil {
		return nil, err
	}

	if !checkHeaderLen(strings.TrimRight(string(headDataByte), HeaderSplit)) {
		return nil, errors.New("header len error")
	}

	bodyDataByte, err := cc.read(DataSplit)
	if err != nil {
		return nil, err
	}

	var check = make([]byte, CheckLen)
	if _, err := cc.bufReadConn.Read(check); err != nil {
		return nil, err
	}

	var allByte []byte
	allByte = append(allByte, headDataByte...)
	allByte = append(allByte, bodyDataByte...)
	if !checkCode(allByte, check) {
		return nil, errors.New("check code error")
	}

	var endTag = make([]byte, len(EndTag))
	if _, err := cc.bufReadConn.Read(endTag); err != nil {
		return nil, err
	}
	if string(endTag) != EndTag {
		return nil, errors.New("end tag not match")
	}

	decoder := mahonia.NewDecoder("GBK")
	m := &NeamMsg{
		Header: strings.TrimRight(decoder.ConvertString(string(headDataByte)), HeaderSplit),
		Data:   strings.TrimRight(decoder.ConvertString(string(bodyDataByte)), DataSplit),
	}

	logrus.Infof("msg :%v", m)

	return m, nil
}

func (cc *clientConn) read(endTag string) ([]byte, error) {
	var (
		buff    = make([]byte, 1)
		allByte []byte
	)

	for {
		if _, err := cc.bufReadConn.Read(buff); err != nil {
			return nil, err
		}
		allByte = append(allByte, buff[0])
		if string(buff[0]) == endTag[0:1] {
			var end = make([]byte, 2)
			if _, err := cc.bufReadConn.Read(end); err != nil {
				return nil, err
			}
			allByte = append(allByte, end...)
			if string(end) == endTag[1:] {
				break
			}
		}
	}

	return allByte, nil
}

func (cc *clientConn) dispatch(rawMsg *NeamMsg) error {

	_ = parseHeader(rawMsg.Header)

	//logrus.Debugf("raw: %+v", raw)

	ackMsg := cc.getAck(rawMsg.Header)

	_, err := cc.bufReadConn.Write(ackMsg)
	if err != nil {
		logrus.Errorf("ack msg error: %v", err)

		return err
	}

	return nil
}

func (cc *clientConn) getAck(header string) []byte {
	headerData := strings.Join([]string{header, HeaderSplit, time.Now().Format(TimeFormat), DataSplit}, "")
	code := computationalCheckCode([]byte(headerData))
	m := strings.Join([]string{headerData, code, EndTag}, "")
	encoder := mahonia.NewEncoder("GBK")

	return []byte(encoder.ConvertString(m))
}
