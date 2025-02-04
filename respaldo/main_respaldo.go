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
	replicatedUsers []BackupUser // Lista de usuarios replicados
	backupMutex     sync.Mutex   // Mutex para acceso concurrente
)

func main() {
	go periodicSync() // Inicia la sincronización periódica con el servidor principal

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

		// Obtener la lista de usuarios del servidor principal
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
		// Actualizamos o eliminamos usuarios según los cambios detectados
		updateAndRemoveUsers(principalUsers)
		backupMutex.Unlock()

		time.Sleep(10 * time.Second) // Intervalo entre sincronizaciones
	}
}

// Actualizar y eliminar usuarios en el servidor de respaldo
func updateAndRemoveUsers(principalUsers []BackupUser) {
	// Mapa para detectar usuarios que ya existen en el principal
	principalMap := make(map[int]BackupUser)
	for _, user := range principalUsers {
		principalMap[user.ID] = user
	}

	// Actualizar usuarios existentes o marcarlos como "presentes"
	presentUsers := make(map[int]bool)
	for i, replicaUser := range replicatedUsers {
		if principalUser, exists := principalMap[replicaUser.ID]; exists {
			// Si el usuario existe, verificar si necesita ser actualizado
			if replicaUser.Name != principalUser.Name || replicaUser.Username != principalUser.Username {
				fmt.Printf("Actualizando usuario: %v\n", principalUser)
				replicatedUsers[i] = principalUser // Actualiza el usuario
			}
			presentUsers[replicaUser.ID] = true
		}
	}

	// Eliminar usuarios que ya no están en el servidor principal
	for i := len(replicatedUsers) - 1; i >= 0; i-- {
		if !presentUsers[replicatedUsers[i].ID] {
			fmt.Printf("Eliminando usuario: %v\n", replicatedUsers[i])
			replicatedUsers = append(replicatedUsers[:i], replicatedUsers[i+1:]...)
		}
	}

	// Agregar nuevos usuarios desde el servidor principal
	for _, principalUser := range principalUsers {
		found := false
		for _, replicaUser := range replicatedUsers {
			if principalUser.ID == replicaUser.ID {
				found = true
				break
			}
		}
		if !found {
			fmt.Printf("Agregando nuevo usuario: %v\n", principalUser)
			replicatedUsers = append(replicatedUsers, principalUser)
		}
	}
}

// Obtener usuarios replicados
func getReplicatedUsers(c *gin.Context) {
	backupMutex.Lock()
	defer backupMutex.Unlock()

	c.JSON(http.StatusOK, replicatedUsers)
}
