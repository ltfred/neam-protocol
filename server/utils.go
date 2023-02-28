package server

import (
	"strconv"
)

// 校验码检查
func checkCode(data []byte, check []byte) bool {
	code := computationalCheckCode(data)

	return code == string(check)
}

// 计算校验码
func computationalCheckCode(data []byte) string {
	var checkReg byte = 0x00
	for _, v := range data {
		checkReg = v ^ checkReg
	}

	code := strconv.FormatInt(int64(checkReg), 16)
	if len(code) == 1 {
		code = "0" + code
	}

	return code
}

// 解析header(header = dataType(4)+mn+time(19)+len(4))
func parseHeader(header string) map[string]string {
	mnLen := len(header) - DataTypeLen - TimeLen - HeaderLenIdentifyLen

	return map[string]string{
		"DataType": header[0:DataTypeLen],
		"MN":       header[4 : DataTypeLen+mnLen],
		"Time":     header[DataTypeLen+mnLen : DataTypeLen+mnLen+TimeLen],
		"HeadLen":  header[DataTypeLen+mnLen+TimeLen:],
	}
}

// 检查header长度
func checkHeaderLen(header string) bool {
	headerLen := len(header)
	l, _ := strconv.ParseInt(header[headerLen-HeaderLenIdentifyLen:headerLen], 16, 64)

	return headerLen == int(l)+HeaderLenIdentifyLen
}
