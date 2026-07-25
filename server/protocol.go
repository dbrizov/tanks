package main

// Input is a client → server message: the player's intent.
type Input struct {
	PlayerId PlayerId `json:"-"`

	AxisX  float64 `json:"ax"` // horizontal move axis [-1, 1]
	AxisY  float64 `json:"ay"` // vertical move axis [-1, 1]
	MouseX float64 `json:"mx"` // mouse x in world coords (aim target)
	MouseY float64 `json:"my"` // mouse y in world coords (aim target)
	Fire   bool    `json:"fire"`
}

type TankState struct {
	Id      PlayerId `json:"id"`
	PosX    float64  `json:"x"`
	PosY    float64  `json:"y"`
	RotBody float64  `json:"rb"`
	RotAim  float64  `json:"ra"`
}

type Snapshot struct {
	Tick  int         `json:"tick"`
	Tanks []TankState `json:"tanks"`
}
