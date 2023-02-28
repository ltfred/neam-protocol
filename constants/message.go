package constants

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
