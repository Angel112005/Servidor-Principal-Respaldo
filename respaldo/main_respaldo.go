package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// Estructura del usuario en el respaldo
type BackupUser struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Username string `json:"username"`
}

var (
	replicatedUsers []BackupUser
	backupMutex     sync.Mutex
)

func main() {
	go periodicSync() // Inicia la sincronización periódica

	r := gin.Default() // Inicializa el router Gin

	// Endpoint para obtener usuarios replicados
	r.GET("/replica", getReplicatedUsers)

	// Inicia el servidor
	r.Run(":8081")
}

// Sincronización periódica con el servidor principal (Short Polling)
func periodicSync() {
	for {
		fmt.Println("Sincronizando con el servidor principal...")
		resp, err := http.Get("http://localhost:8080/users")
		if err != nil {
			fmt.Println("Error al conectar con el servidor principal:", err)
			time.Sleep(5 * time.Second)
			continue
		}
		defer resp.Body.Close()

		var principalUsers []BackupUser
		err = json.NewDecoder(resp.Body).Decode(&principalUsers)
		if err != nil {
			fmt.Println("Error al procesar los datos del servidor principal:", err)
			time.Sleep(5 * time.Second)
			continue
		}

		backupMutex.Lock()
		// Identificar y replicar únicamente los nuevos usuarios
		for _, principalUser := range principalUsers {
			found := false
			for _, replicaUser := range replicatedUsers {
				if principalUser.ID == replicaUser.ID {
					found = true
					break
				}
			}
			if !found {
				replicatedUsers = append(replicatedUsers, principalUser)
				fmt.Printf("Usuario replicado: %v\n", principalUser)
			}
		}
		backupMutex.Unlock()

		time.Sleep(10 * time.Second)
	}
}

// Obtener usuarios replicados
func getReplicatedUsers(c *gin.Context) {
	backupMutex.Lock()
	defer backupMutex.Unlock()

	c.JSON(http.StatusOK, replicatedUsers)
}
