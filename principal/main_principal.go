package main

import (
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
)

// Estructura del usuario
type User struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

var (
	users []User
	mutex sync.Mutex
)

func main() {
	r := gin.Default() // Inicializa el router Gin

	// Rutas para el CRUD
	r.GET("/users", getUsers)
	r.POST("/users", createUser)
	r.PUT("/users/:id", updateUser)
	r.DELETE("/users/:id", deleteUser)

	// Ruta para long-polling
	r.GET("/long-poll", longPollUsers)

	// Iniciar servidor
	r.Run(":8080") // Inicia el servidor en el puerto 8080
}

// Obtener todos los usuarios
func getUsers(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	c.JSON(http.StatusOK, users)
}

// Crear un nuevo usuario
func createUser(c *gin.Context) {
	var newUser User
	if err := c.ShouldBindJSON(&newUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	newUser.ID = len(users) + 1
	users = append(users, newUser)

	c.JSON(http.StatusCreated, newUser)
}

// Actualizar un usuario existente
func updateUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	var updatedUser User
	if err := c.ShouldBindJSON(&updatedUser); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Datos inválidos"})
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for i, user := range users {
		if user.ID == id {
			users[i] = updatedUser
			c.JSON(http.StatusOK, updatedUser)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
}

// Eliminar un usuario
func deleteUser(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID inválido"})
		return
	}

	mutex.Lock()
	defer mutex.Unlock()

	for i, user := range users {
		if user.ID == id {
			users = append(users[:i], users[i+1:]...)
			c.JSON(http.StatusOK, gin.H{"message": "Usuario eliminado"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Usuario no encontrado"})
}

// Long polling para enviar usuarios uno por uno
func longPollUsers(c *gin.Context) {
	mutex.Lock()
	defer mutex.Unlock()

	for _, user := range users {
		c.JSON(http.StatusOK, user)
	}
}
