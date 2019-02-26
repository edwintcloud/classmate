package attendance

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/globalsign/mgo/bson"
	"golang.org/x/crypto/bcrypt"
)

// Create a new Person with default role of student
func (p *Person) Create() (err error) {

	// assign role of student
	p.Role = "student"

	// TODO: validation

	// hash password
	password, err := bcrypt.GenerateFromPassword([]byte(p.Password), 6)
	if err != nil {
		return err
	}
	p.Password = string(password)

	// create new person in db
	return db.persons.Insert(&p)

}

// Find finds a person by id or email depending on if id is set
func (p *Person) Find() error {

	if bson.IsObjectIdHex(p.ID.Hex()) {
		// find by id
		return db.persons.FindId(p.ID).One(&p)
	}

	// else find by email
	return db.persons.Find(bson.M{"email": p.Email}).One(&p)
}

// Authenticate authenticates a person an generates an authorization jwt
func (p *Person) Authenticate(password string) error {

	// find person by email
	err := p.Find()
	if err != nil {
		return err
	}

	// ensure password matches
	err = bcrypt.CompareHashAndPassword([]byte(p.Password), []byte(password))
	if err != nil {
		return err
	}

	// generate jwt token
	token := jwt.New(jwt.SigningMethodHS256)

	// set jwt claims
	claims := token.Claims.(jwt.MapClaims)
	claims["id"] = p.ID.Hex()
	claims["exp"] = time.Now().Add(time.Hour * 72).Unix()

	// set person token to generated jwt
	p.Token, err = token.SignedString(s.JwtSecret)

	return err

}

// Create a class
func (c *Class) Create() error {
	return db.classes.Insert(&c)
}

// Find a class by _id
func (c *Class) Find() error {
	return db.classes.FindId(c.ID).One(&c)
}
