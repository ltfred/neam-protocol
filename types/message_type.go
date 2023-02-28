package types

type MessageType string

const (
	// 监测数据
	MessageTypeBN01 MessageType = "bn01" // 30 秒值 标况
	MessageTypeJR01 MessageType = "JR01" // 30 秒值 实况
	MessageTypeJZ12 MessageType = "JZ12" // 5 分钟均值 标况
	MessageTypeJR12 MessageType = "JR12" // 5 分钟均值 实况
	MessageTypeJZ16 MessageType = "JZ16" // 1 小时均值 标况
	MessageTypeJR16 MessageType = "JR16" // 1 小时均值 实况
	MessageTypeJZ18 MessageType = "JZ18" // AQI日均值 标况
	MessageTypeJR18 MessageType = "JR18" // AQI日均值 实况
	MessageTypeJZ06 MessageType = "JZ06" // API日均值 标况
	MessageTypeJR06 MessageType = "JR06" // API日均值 实况
	MessageTypeJZ24 MessageType = "JZ24" // 气象日均值 标况
	MessageTypeJZ25 MessageType = "JZ25" // 温室气体日均值 标况
	MessageTypeJZ31 MessageType = "JZ31" // 动环实时数据
	MessageTypeJZ33 MessageType = "JZ33" // 动环小时数据
)
