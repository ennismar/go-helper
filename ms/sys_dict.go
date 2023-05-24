package ms

type SysDict struct {
	M
	Name      string        `gorm:"index:idx_name_unique,unique;comment:name" json:"name"`
	Desc      string        `gorm:"comment:description" json:"desc"`
	Status    *uint         `gorm:"type:tinyint(1);default:1;comment:status(0: disabled, 1: enabled)" json:"status"`
	DictDatas []SysDictData `gorm:"foreignKey:DictId;comment:one2many datas" json:"dictDatas"`
}

type SysDictData struct {
	M
	Key      string  `gorm:"comment:key" json:"key"`
	Val      string  `gorm:"comment:val" json:"val"`
	Addition string  `gorm:"comment:custom addition params" json:"addition"`
	Sort     *uint   `gorm:"comment:sort" json:"sort"`
	Status   *uint   `gorm:"type:tinyint(1);default:1;comment:status(0: disabled, 1: enabled)" json:"status"`
	DictId   uint    `gorm:"comment:dict id" json:"dictId"`
	Dict     SysDict `gorm:"foreignKey:DictId" json:"dict"`
}
