package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"

	"ttb-bluebook/internal/database"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Report struct {
	Total string `json:"total"`
	Model string `json:"model"`
	Gear  string `json:"gear"`
}

func main() {
	r := gin.New()

	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"POST", "PUT", "PATCH", "DELETE"},
		AllowHeaders: []string{"Content-Type,access-control-allow-origin, access-control-allow-headers"},
	}))

	db, err := database.Database()
	if err != nil {
		panic(err)
	}

	r.GET("/cars/report", func(c *gin.Context) {
		var carReport []Report
		db.Raw("SELECT COUNT(model) as total, model, gear FROM catalogs GROUP BY model, gear ORDER BY COUNT(model) DESC").Scan(&carReport)
		c.JSON(http.StatusOK, carReport)
	})

	r.GET("/cars", func(c *gin.Context) {
		var cars []database.Catalog
		db.Find(&cars)
		c.JSON(http.StatusOK, cars)
	})

	r.GET("/cars/:id", func(c *gin.Context) {
		carId := c.Param("id")
		var car database.Catalog
		result := db.First(&car, "id = ?", carId)
		if result.RowsAffected < 1 {
			c.JSON(http.StatusNotFound, gin.H{
				"message": "Not Found Car",
			})
			return
		}

		c.JSON(http.StatusOK, car)
	})

	r.DELETE("/cars/:id", func(c *gin.Context) {
		var car database.Catalog
		carId := c.Param("id")
		removed := db.Delete(&car, carId)
		if removed.Error != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"message": removed.Error.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{})
	})

	r.POST("/cars", func(c *gin.Context) {
		var car database.Catalog
		err := c.ShouldBindJSON(&car)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		db.Create(&car)
		c.JSON(http.StatusOK, car)
	})

	r.POST("/cars/upload", func(c *gin.Context) {

		file, header, err := c.Request.FormFile("upload")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"message": err.Error(),
			})
			return
		}

		filename := header.Filename
		out, err := os.Create("./images/" + filename)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
		}
		defer out.Close()
		_, err = io.Copy(out, file)
		if err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, gin.H{
			"keyImage": filename,
		})
	})

	r.GET("/cars/image/:file", func(c *gin.Context) {
		fileName := c.Param("file")
		c.File("./images/" + fileName)
	})

	r.PUT("/cars/:id", func(c *gin.Context) {
		var newCars database.Catalog

		err := c.ShouldBindJSON(&newCars)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"message": err.Error(),
			})
			return
		}

		values := reflect.ValueOf(newCars)
		types := values.Type()
		dataObj := make(map[string]interface{})
		responseData := make(map[string]interface{})
		for i := 0; i < values.NumField(); i++ {
			if values.Field(i).String() != "" {
				jsonField := types.Field(i).Tag.Get("json")
				field := types.Field(i).Name
				responseData[jsonField] = values.Field(i).Interface()
				dataObj[field] = values.Field(i).Interface()
			}
		}

		carIdParam := c.Param("id")
		carId, err := strconv.ParseUint(carIdParam, 10, 64)
		if err != nil {
			panic(err)
		}

		db.Model(&newCars).Where("id = ?", carId).Updates(dataObj)
		c.JSON(http.StatusOK, responseData)
	})

	r.Run(":9000") // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}
