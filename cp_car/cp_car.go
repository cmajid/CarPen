package cp_car

import (
	"image"
	"image/color"
	"math"

	"github.com/cmajid/carpen/cp_pivot"
	"github.com/cmajid/carpen/cp_vector"
	"github.com/cmajid/carpen/cp_wheel"
	"github.com/fogleman/gg"
)

type Point struct {
	X, Y float64
}

type Car struct {
	RotateLeft,
	RotateRight,
	Accelerate,
	Decelerate bool
	WheelWidth,
	WheelHeight,
	Width,
	Height int
	WheelRotationStep,
	WheelMaxAngle,
	WheelAngle,
	X,
	Y,
	Rotation,
	Speed,
	Acceleration float64
	Pivot          cp_pivot.Pivot
	FrontPivot     cp_pivot.FrontPivot
	RearPivot      cp_pivot.RearPivot
	DirectionPivot cp_pivot.DirectionPivot
	RearPivotAbs   cp_pivot.RearPivotAbs
	TempDirPivot   cp_pivot.TempDirPivot
	Wheels         []cp_wheel.Wheel
	Direction      cp_pivot.Direction
}

func (c *Car) UpdateDirection() {
	v1 := cp_vector.Vector{X: c.DirectionPivot.X - c.FrontPivot.X, Y: c.DirectionPivot.Y - c.FrontPivot.Y}
	c.Direction = v1.Normalize()
	c.Direction.X *= c.Speed
	c.Direction.Y *= c.Speed
}

func init() {

}
func (p Point) Length() float64 {

	return math.Sqrt(math.Pow(p.X, 2) + math.Pow(p.Y, 2))
}
func (car *Car) DrawCar() image.Image {

	dc := gg.NewContext(640, 480)
	car.DrawWheels(dc)
	dc.Translate(car.Pivot.X, car.Pivot.Y)
	dc.Rotate(car.Rotation * math.Pi / 180)
	dc.SetColor(color.RGBA{0xff, 0, 0, 0xff})
	dc.DrawRectangle(-50, -20, float64(car.Width), float64(car.Height))
	dc.Fill()
	return dc.Image()

	//op := &ebiten.DrawImageOptions{}
	//op.GeoM.Translate(-float64(car.width)/2, -float64(car.height)/2)
	//op.GeoM.Translate(car.pivot.X-float64(car.width/2), car.pivot.Y)

	// op.GeoM.Translate(car.pivot.X, car.pivot.Y)
	// //	op.GeoM.Translate(-float64(car.width)/2, -float64(car.height)/2)
	// op.GeoM.Rotate(car.rotation * math.Pi / 180)

	// carImage := ebiten.NewImage(car.width, car.height)
	// carImage.Fill(color.Black)

	//op.GeoM.Translate(-50, -20)
	//dst.DrawImage(carImage, op)

	//fmt.Printf("dst: %v\n", car.pivot)

	// ebitenutil.DrawLine(dst, car.frontPivot.X-50, car.frontPivot.Y, car.frontPivot.X+50, car.frontPivot.Y, color.White)
	// ebitenutil.DrawLine(dst, car.rearPivot.X-50, car.rearPivot.Y, car.rearPivot.X+50, car.rearPivot.Y, color.White)
	// ebitenutil.DrawLine(dst, car.frontPivot.X, car.frontPivot.Y, car.rearPivot.X, car.rearPivot.Y, color.White)
	// ebitenutil.DrawLine(dst, car.frontPivot.X, car.frontPivot.Y, car.tempDirPivot.X, car.tempDirPivot.Y, color.White)

	//fmt.Printf("dst:%v %v\n", car.directionPivot.X, car.directionPivot.Y)
	//	ebitenutil.DrawRect(dst, car.directionPivot.X, car.directionPivot.Y, 2, 2, color.White)

}
func (car *Car) DrawWheels(dc *gg.Context) gg.Context {

	for i := 0; i < len(car.Wheels); i++ {
		if car.RotateLeft {
			if car.WheelAngle > -car.WheelMaxAngle {
				car.WheelAngle = car.WheelAngle - car.WheelRotationStep

			}
		}
		if car.RotateRight {
			if car.WheelAngle < car.WheelMaxAngle {
				car.WheelAngle = car.WheelAngle + car.WheelRotationStep
			}
		}
		dc.Push()
		dc.Translate(car.Pivot.X, car.Pivot.Y)
		dc.Rotate(car.Rotation * math.Pi / 180)
		var o = car.Wheels[i]
		dc.Translate(o.X, o.Y)
		if i < 2 {
			dc.Rotate((car.WheelAngle) * math.Pi / 180)
		}
		dc.SetRGB(0, 0, 0)
		dc.DrawRectangle(-6, -15, float64(car.WheelWidth), float64(car.WheelHeight))
		dc.Fill()
		dc.Pop()

	}
	car.DirectionPivot.X = 50*math.Cos((car.WheelAngle+car.Rotation-90)*math.Pi/180) + car.FrontPivot.X
	car.DirectionPivot.Y = 50*math.Sin((car.WheelAngle+car.Rotation-90)*math.Pi/180) + car.FrontPivot.Y

	car.TempDirPivot.X = 50*math.Cos((car.WheelAngle-90)*math.Pi/180) + car.FrontPivot.X
	car.TempDirPivot.Y = 50*math.Sin((car.WheelAngle-90)*math.Pi/180) + car.FrontPivot.X

	car.UpdateDirection()
	return *dc
}

func (car *Car) Move() error {

	_flag := false
	if car.Accelerate && car.Speed < 6 {
		car.Speed += car.Acceleration
		_flag = true
	} else if car.Speed > 0 {
		car.Speed -= car.Acceleration
		_flag = true
	}
	if car.Decelerate && car.Speed > -3 {
		car.Speed -= car.Acceleration
		_flag = true
	} else if car.Speed < -0.3 {
		car.Speed += car.Acceleration
		_flag = true
	}

	if !_flag {
		car.Speed = 0
	}
	car.Pivot.X += car.Direction.X
	car.Pivot.Y += car.Direction.Y

	//fmt.Printf("speed: %v\n", car.speed)
	//fmt.Printf("acceleration: %v\n", car.acceleration)

	// Drift!!! section :D
	v := cp_vector.Vector{X: car.Pivot.X - car.RearPivotAbs.X, Y: car.Pivot.Y - car.RearPivotAbs.Y}
	var rotation = math.Atan2(-v.Y, v.X) * 180 / math.Pi

	rotation += 180
	rotation = 360 - rotation - 90
	car.Rotation = rotation

	car.RearPivotAbs = cp_pivot.RearPivotAbs{
		X: 160*math.Cos((car.Rotation+90)*math.Pi/180) + car.Pivot.X,
		Y: 160*math.Sin((car.Rotation+90)*math.Pi/180) + car.Pivot.Y,
	}

	//fmt.Printf("rearPivotAbs: %v\n", car.rearPivotAbs)

	return nil
}
