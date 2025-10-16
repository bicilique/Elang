package entity

type Framework struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func (Framework) TableName() string {
	return "framework"
}
