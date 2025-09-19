package main

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// вынести логику бд в отдельный файл
const (
	port = ":8080"
)

type Frame struct {
	Id    int    `json:"id"`
	Date  string `json:"date"`
	Image string `json:"image"`
}

func main() {

	router := gin.Default() // роутер

	//прописываем хендлеры
	router.GET("/frames", getAllFrames)
	router.GET("/frames/:id", getFrameById)
	router.POST("/frames/create", createFrame)
	router.DELETE("/frames/delete/:id", deleteFrameById)
	router.PUT("frames/update/:id", updateFrameById)

	//выдача статических изображений для фронтенда
	router.Static("/images", "/var/data")

	router.Run(port)
}

// получаем инфу о всех кадрах в json
func getAllFrames(c *gin.Context) {
	//создаем коннект с бд
	connInfo := "host=localhost port=5432 user=postgres password=postgres dbname=frames sslmode=disable"
	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//читаем из бд в слайс структур
	frames := []Frame{}
	rows, err := db.Query("SELECT * FROM frames")
	if err != nil {
		panic(err)
	}

	//пишем в результаты в слайс
	for rows.Next() {
		var f Frame
		err = rows.Scan(&f.Id, &f.Date, &f.Image)
		if err != nil {
			panic(err)
		}

		frames = append(frames, f)
	}

	for _, val := range frames {
		fmt.Println(val.Id, val.Date, val.Image)
	}

	//пишем в body reponse json кадров
	c.IndentedJSON(http.StatusOK, frames)

}

// получаем инфу об одном кадре
func getFrameById(c *gin.Context) {

	//считали id кадра
	id := c.Param("id")

	connInfo := "host=localhost port=5432 user=postgres password=postgres dbname=frames sslmode=disable"
	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//запрос к бд
	var f Frame
	res := db.QueryRow("SELECT * FROM frames WHERE id=$1", id)
	res.Scan(&f.Id, &f.Date, &f.Image)
	if f.Id == 0 {
		c.String(http.StatusNotFound, "Нет такого кадра")
		return
	}
	//записали результаты в структуру и ее прокинули в response body
	c.IndentedJSON(http.StatusOK, f)
}

// создаем кадр инфу принимаем в json
func createFrame(c *gin.Context) {
	var f Frame
	var id int
	//считали из request body в структуру
	if err := c.BindJSON(&f); err != nil {
		c.String(http.StatusBadRequest, "Неверно передана информация о кадре")
		return
	}

	connInfo := "host=localhost port=5432 user=postgres password=postgres dbname=frames sslmode=disable"
	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//вставляем запись о кадре в бд
	err = db.QueryRow("INSERT INTO frames (date, image) VALUES ($1, $2) returning id", f.Date, f.Image).Scan(&id)
	if err != nil {
		panic(err)
	}

	f.Id = id
	//вернули созданный кадр
	c.IndentedJSON(http.StatusCreated, f)
}

// удаляем кадр по id
func deleteFrameById(c *gin.Context) {
	id := c.Param("id")

	connInfo := "host=localhost port=5432 user=postgres password=postgres dbname=frames sslmode=disable"
	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	//удаляем запись о кадре по id
	res, err := db.Exec("DELETE FROM frames WHERE id=$1", id)
	if err != nil {
		panic(err)
	}

	fmt.Println(res.RowsAffected())
	c.String(http.StatusOK, "Кадр удален")
}

// обновляем кадр json из req body + id из url
func updateFrameById(c *gin.Context) {
	id := c.Param("id")

	var f Frame
	//считали из тела запроса в структуру
	if err := c.BindJSON(&f); err != nil {
		return
	}

	connInfo := "host=localhost port=5432 user=postgres password=postgres dbname=frames sslmode=disable"
	db, err := sql.Open("postgres", connInfo)
	if err != nil {
		panic(err)
	}
	defer db.Close()
	//обновляем инфу в базе данных
	res, err := db.Exec("UPDATE frames SET date=$1, image=$2 WHERE id=$3", f.Date, f.Image, id)
	if err != nil {
		panic(err)
	}

	fmt.Println(res.RowsAffected())
	//возвращаем инфу об обновленном кадре
	c.IndentedJSON(http.StatusOK, f)
}
