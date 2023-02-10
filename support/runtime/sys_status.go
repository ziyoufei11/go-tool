package runtime

type Status int

const (
	SysOK               Status = 0     //状态OK
	SysUnderMaintenance Status = 12000 //维护中
	SysUnConfigure      Status = 12100 //未完成配置
	SysMySQLError       Status = 12200 //MySQL故障
	SysRedisError       Status = 12300 //Redis故障
)

var SysStatus Status

var msg = map[Status]string{
	SysOK:               "",
	SysUnderMaintenance: "系统维护中",
	SysUnConfigure:      "未完成配置",
	SysMySQLError:       "MySQL维护中",
	SysRedisError:       "Redis维护中",
}

func SysStatusMsg(code Status) string {
	s, ok := msg[code]
	if !ok {
		return "未知错误"
	}

	return s
}
