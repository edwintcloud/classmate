package attendance

import (
	"time"

	"github.com/globalsign/mgo/bson"
)

// Class is our class model
type Class struct {
	ID         bson.ObjectId   `json:"_id,omitempty" bson:"_id,omitempty"`
	Title      string          `json:"title" bson:"title"`
	Instructor bson.ObjectId   `json:"instructor" bson:"instructor"`
	StartTime  time.Time       `json:"start_time" bson:"start_time"`
	EndTime    time.Time       `json:"end_time" bson:"end_time"`
	StartDate  time.Time       `json:"start_date" bson:"start_date"`
	EndDate    time.Time       `json:"end_date" bson:"end_date"`
	Location   string          `json:"location" bson:"location"`
	Students   []bson.ObjectId `json:"students" bson:"students"`
}

// Person is our student and instructor model
type Person struct {
	ID        bson.ObjectId   `json:"-" bson:"_id,omitempty"`
	Email     string          `json:"email" bson:"email"`
	Password  string          `json:"password,omitempty" bson:"password"`
	FirstName string          `json:"first_name" bson:"first_name"`
	LastName  string          `json:"last_name" bson:"last_name"`
	Role      string          `json:"-" bson:"role"`
	Token     string          `json:"token,omitempty" bson:"-"`
	Classes   []bson.ObjectId `json:"classes" bson:"classes"`
}
