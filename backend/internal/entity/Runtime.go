package entity

type Runtime struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}

func (Runtime) TableName() string {
	return "runtime"
}
