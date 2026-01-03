package model

// RedisserviceconfigTb 数据模型
type RedisserviceconfigTb struct {
	Autoid int64 `xorm:"pk autoincr int notnull 'AutoId'" json:"auto_id"`
	Groupname string `xorm:"varchar(50) 'GroupName'" json:"group_name"`
	Configkey string `xorm:"varchar(50) 'ConfigKey'" json:"config_key"`
	Configvalue string `xorm:"varchar(100) 'ConfigValue'" json:"config_value"`
	Remark string `xorm:"varchar(150) 'Remark'" json:"remark"`
	State string `xorm:"tinyint unsigned notnull 'State'" json:"state"`
}

// TableName returns the table name
func (RedisserviceconfigTb) TableName() string {
	return "redisserviceconfig_tb"
}
