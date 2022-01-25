package handler

import (
	"aws-iam/entity"
	service "aws-iam/usecase/myobject"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
)

func PushObject(c *gin.Context) {
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var object entity.MyObject
	err = json.Unmarshal(body, &object)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = service.PushObject(&object)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, object)
}

func GetObject(c *gin.Context) {
	key := c.Param("key")
	object, err := service.GetObject(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, object)
}

func ListObjects(c *gin.Context) {
	fromCacheStr := c.Query("cache")
	fromCache, err := strconv.ParseBool(fromCacheStr)
	if err != nil {
		fromCache = true
	}

	objects, err := service.ListObjects(fromCache)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, objects)
}
