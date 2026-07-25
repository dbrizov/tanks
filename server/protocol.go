package main

// Input is a client → server message: the player's intent.
type Input struct {
	PlayerId PlayerId `json:"-"`

	Ax   float64 `json:"ax"`   // horizontal move axis [-1, 1]
	Ay   float64 `json:"ay"`   // vertical move axis [-1, 1]
	Fire bool    `json:"fire"`
}

type TankState struct {
	Id    PlayerId `json:"id"`
	X     float64  `json:"x"`
	Y     float64  `json:"y"`
	Angle float64  `json:"angle"`
}

type Snapshot struct {
	Tick  int         `json:"tick"`
	Tanks []TankState `json:"tanks"`
}
