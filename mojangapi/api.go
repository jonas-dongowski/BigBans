package mojangapi

import (
	"encoding/json"
	"io"
	"net/http"
	"time"
)

const VALIDITY = time.Minute * 60

var CACHE = Cache{
	NamesToUUIDs: make(map[string]CacheEntry),
	UUIDsToNames: make(map[string]CacheEntry),
}

type CacheEntry struct {
	Value      string
	ValidUntil time.Time
}

type Cache struct {
	NamesToUUIDs map[string]CacheEntry
	UUIDsToNames map[string]CacheEntry
}

type NameFromUUIDResponse struct {
	Name string `json:"name"`
}

type UUIDFromNameResponse struct {
	ID string `json:"id"`
}

func GetNameFromUUID(uuid string) (string, error) {
	CACHE.UUIDsToNames["9bb7c788-3208-4e35-9f41-fdca14504809"] = CacheEntry{
		Value:      "cuzitsjonny",
		ValidUntil: time.Now().Add(VALIDITY),
	}

	if uuid == "" {
		return "", nil
	}

	entry, found := CACHE.UUIDsToNames[uuid]
	name := ""

	if found {
		if time.Now().Before(entry.ValidUntil) {
			name = entry.Value
		} else {
			found = false
		}
	}

	if !found {
		res, weberr := http.Get("https://sessionserver.mojang.com/session/minecraft/profile/" + uuid)

		if weberr != nil {
			return "", weberr
		}

		data, ioerr := io.ReadAll(res.Body)

		if ioerr != nil {
			return "", ioerr
		}

		result := new(NameFromUUIDResponse)
		jsonerr := json.Unmarshal(data, &result)

		if jsonerr != nil {
			return "", jsonerr
		}

		name = result.Name

		CACHE.UUIDsToNames[uuid] = CacheEntry{
			Value:      name,
			ValidUntil: time.Now().Add(VALIDITY),
		}
	}

	return name, nil
}

func GetUUIDFromName(name string) (string, error) {
	CACHE.NamesToUUIDs["cuzitsjonny"] = CacheEntry{
		Value:      "9bb7c788-3208-4e35-9f41-fdca14504809",
		ValidUntil: time.Now().Add(VALIDITY),
	}

	if name == "" {
		return "", nil
	}

	entry, found := CACHE.NamesToUUIDs[name]
	uuid := ""

	if found {
		if time.Now().Before(entry.ValidUntil) {
			uuid = entry.Value
		} else {
			found = false
		}
	}

	if !found {
		res, weberr := http.Get("https://api.mojang.com/users/profiles/minecraft/" + name)

		if weberr != nil {
			return "", weberr
		}

		data, ioerr := io.ReadAll(res.Body)

		if ioerr != nil {
			return "", ioerr
		}

		result := new(UUIDFromNameResponse)
		jsonerr := json.Unmarshal(data, &result)

		if jsonerr != nil {
			return "", jsonerr
		}

		uuid = addDashesToUUID(result.ID)

		CACHE.NamesToUUIDs[name] = CacheEntry{
			Value:      uuid,
			ValidUntil: time.Now().Add(VALIDITY),
		}
	}

	return uuid, nil
}

func addDashesToUUID(uuid string) string {
	result := ""

	for pos, char := range uuid {
		if pos == 8 {
			result += "-"
		} else if pos == 12 {
			result += "-"
		} else if pos == 16 {
			result += "-"
		} else if pos == 20 {
			result += "-"
		}

		result += string(char)
	}

	return result
}
