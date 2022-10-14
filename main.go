package main

import (
	"fmt"
	"strconv"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	jwtware "github.com/gofiber/jwt/v3"
	"github.com/golang-jwt/jwt"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
)

var db *sqlx.DB

const jwtSecret = "nonza"

func main(){

	var err error
	db, err = sqlx.Open("mysql","root:non256199@tcp(127.0.0.1:3306)/user_go")
	if err != nil{
		panic(err)
	}
	app := fiber.New()
	app.Use("/hello",jwtware.New(jwtware.Config{
		SigningMethod: "HS256",
		SigningKey: []byte(jwtSecret),
		SuccessHandler: func(c *fiber.Ctx) error {
			return c.Next()
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return fiber.ErrUnauthorized
		},
	}))
	app.Post("/signup",SignUp)
	app.Post("/hello",Hello)
	app.Post("/login",Login)

	app.Listen(":8000")


}
func SignUp(c *fiber.Ctx)error{
	//get request body
	request := SignupRequest{}
	err :=c.BodyParser(&request)
	if err != nil{
		return err
	}

	if request.Username == ""||request.Password == "" {
		return fiber.ErrUnprocessableEntity
	}

	password,err := bcrypt.GenerateFromPassword([]byte(request.Password),10)
	if err != nil{
		return fiber.NewError(fiber.StatusUnprocessableEntity,err.Error())
	}

	query := "insert users(username,password) values (?,?)"
	result, err := db.Exec(query,request.Username,string(password))
	if err != nil{
		return fiber.NewError(fiber.StatusUnprocessableEntity,err.Error())
	}

	id, err := result.LastInsertId()
	if err != nil{
		return fiber.NewError(fiber.StatusUnprocessableEntity,err.Error())
	}
	user := User{
		Id: int(id),
		Username: request.Username,
		Password: request.Password,
	}
	return c.Status(fiber.StatusCreated).JSON(user)
}

func Login(c *fiber.Ctx)error{
	request := LoginRequest{}
	err := c.BodyParser(&request)
	if err != nil{
		return err
	}

	if request.Username == ""||request.Password == "" {
		return fiber.ErrUnprocessableEntity
	}

	user := User{}
	query := "select id, username, password from users where username=?"
	err = db.Get(&user,query,request.Username)
	if err != nil{
		return fiber.NewError(fiber.StatusNotFound,"Incorrect username or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password),[]byte(request.Password))
	if err != nil{
		return fiber.NewError(fiber.StatusNotFound,"Incorrect username or password")
	}

	cliams := jwt.StandardClaims{
		Issuer:strconv.Itoa(user.Id),
		ExpiresAt: time.Now().Add(time.Hour*24).Unix(),
	}
	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256,cliams)
	token, err := jwtToken.SignedString([]byte(jwtSecret))
	if err != nil{
		return fiber.ErrInternalServerError
	}

	return c.JSON(fiber.Map{
		"jwtToken":token,
	})

}

func Hello(c *fiber.Ctx) error {
	return c.SendString("Hello world")

}

type User struct{
	Id int `db:"id" json:"id"`
	Username string `db:"username" json:"username"`
	Password string `db:"password" json:"password"`
}

type SignupRequest struct{
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginRequest struct{
	Username string `json:"username"`
	Password string `json:"password"`
}
func Fiber(){
	app := fiber.New(fiber.Config{
		Prefork: true,
	})
	//Middleware
	app.Use("/hello",func(c *fiber.Ctx) error  {
		//Send data to other func from middle using Local()
		c.Locals("non","za")
		fmt.Println("before")
		err := c.Next()
		fmt.Println("after")
		return err
	})

	app.Use(logger.New(logger.Config{
		TimeZone: "Asia/Tokyo",
	}))

	app.Use(requestid.New())

	app.Use(cors.New())

	app.Get("/hello",func(c *fiber.Ctx) error {
		fmt.Println("Hello")
		//Get value from middleware
		name := c.Locals("non")
		return c.SendString(fmt.Sprintf("Get: Hello %v",name))
	})


	app.Post("/hello",func(c *fiber.Ctx) error {
		fmt.Println("Hello")
		return c.SendString("Post: hello non")
	})

	//Parameter
	app.Get("/hello/:name/:surname",func(c *fiber.Ctx) error {
		name:=c.Params(("name"))
		surname:= c.Params("surname")
		return c.SendString("name: "+ name+"surname: "+surname)
	})

	//ParamsInt
	app.Get("/hello/:id",func(c *fiber.Ctx) error {
		id, err := c.ParamsInt("id")
		if err!=nil{
			return fiber.ErrBadRequest
		}
		return c.SendString(fmt.Sprintf("ID: %d",id))
	})

	//Query
	//curl "localhost:8000/query?name=non"
	//curl "localhost:8000/query?name=non&surname=za"
	app.Get("/query",func(c *fiber.Ctx) error {
		name := c.Query("name")
		surname := c.Query("surname")
		return c.SendString("name: "+ name + "surname: " + surname)
	})

	//QueryParser
	// curl "localhost:8000/query2?id=1&name=non"  
	app.Get("/query2",func(c *fiber.Ctx) error {
		person := Person{}
		c.QueryParser(&person)
		return c.JSON((person))
	})

	//POST
	//curl localhost:8000/body -d "{"name":"non"}" -H content-type:application/json -> command to set res in json type

	app.Post("/body",func(c *fiber.Ctx) error {
		fmt.Printf("IsJson: %v\n",c.Is("json"))
		fmt.Println(string(c.Body()))
		return nil
	})
	//Read response and save to struct type
	app.Post("/body",func(c *fiber.Ctx) error {
		fmt.Printf("IsJson: %v\n",c.Is("json"))
		fmt.Println(string(c.Body()))
		person := Person{}
		err :=c.BodyParser(&person)
		if err!= nil{
			return err
		}
		fmt.Println(person)
		return nil

	})

	//Tip:if data in json is dynamic, we can read response and save to struct by using map
	//.We will map string to interface not specific type of data
	app.Post("/body2",func(c *fiber.Ctx) error {
		fmt.Printf("IsJson: %v\n",c.Is("json"))
		data := map[string]interface{}{}
		err := c.BodyParser(&data)
		if err != nil{
			return err
		}
		fmt.Println(data)
		return nil
	})

	//Static file
	app.Static("/","./root")

	//Newerror
	app.Get("/error",func(c *fiber.Ctx) error {
		fmt.Println("error")
		return fiber.NewError(fiber.StatusNotFound,"content not found")
	})

	//Group
	v1 := app.Group("/v1",func(c *fiber.Ctx) error {
		c.Set("Version","v1")
		return c.Next()
	})

	v1.Get("/hello",func(c *fiber.Ctx) error {
		return c.SendString("Hello V1")
	})

	v2 := app.Group("/v2",func(c *fiber.Ctx) error {
		c.Set("Version","v2")
		return c.Next()
	})

	v2.Get("/hello",func(c *fiber.Ctx) error {
		return c.SendString("Hello v2")
	})

	//Mount
	userApp := fiber.New()
	userApp.Get("/login",func(c *fiber.Ctx) error {
		return c.SendString("Login")
	})

	app.Mount("/user",userApp)

	//Serer config
	app.Server().MaxConnsPerIP = 1
	app.Get("/server",func(c *fiber.Ctx) error {
		time.Sleep(time.Second*30)
		return c.SendString("server")
	})

	app.Get("/env",func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"BaseURL": c.BaseURL(),
			"Hostname": c.Hostname(),
			"IP": c.IP(),
			"IPs": c.IPs(),
			"path": c.Path(),
			"Protocol": c.Protocol(),
		})
	})



	app.Listen(":8000")
}

type Person struct{
	Id int `json:"id"`
	Name string `json:"name"`

}
