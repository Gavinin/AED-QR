package services

import (
	"AED-QR/internal/model"
	"sync"
)

var (
	// Map UUID -> Vehicle
	vehicleCache = make(map[string]model.Vehicle)
	cacheMutex   sync.RWMutex
)

// InitCache populates the cache with the provided list of vehicles
func InitCache(vehicles []model.Vehicle) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	// Clear existing cache just in case
	vehicleCache = make(map[string]model.Vehicle)

	for _, v := range vehicles {
		if v.QRCode.UUID != "" {
			vehicleCache[v.QRCode.UUID] = v
		}
	}
}

// UpdateVehicleCache updates or adds a vehicle to the cache
func UpdateVehicleCache(v model.Vehicle) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	if v.QRCode.UUID != "" {
		vehicleCache[v.QRCode.UUID] = v
	}
}

// RemoveVehicleFromCache removes a vehicle from the cache by UUID
// Since we might only have ID during delete, this might be tricky if we don't know the UUID.
// However, in the handler we can look it up before delete.
func RemoveVehicleFromCache(uuid string) {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	delete(vehicleCache, uuid)
}

// GetVehicleByUUID retrieves a vehicle from the cache by UUID
func GetVehicleByUUID(uuid string) (model.Vehicle, bool) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()
	v, ok := vehicleCache[uuid]
	return v, ok
}
