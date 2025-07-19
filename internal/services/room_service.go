package services

import (
"errors"
"sync"
"github.com/johnpr01/home-automation/internal/models"
)


type RoomService struct {
	rooms map[string]*models.Room
	mutex sync.RWMutex
}

func NewRoomService() *RoomService {
	return &RoomService{
		rooms: make(map[string]*models.Room),
	}
}

func (rs *RoomService) GetRoom(id string) (*models.Room, error) {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	
	room, exists := rs.rooms[id]
	if !exists {
		return nil, errors.New("room not found")
	}
	return room, nil
}

func (rs *RoomService) AddRoom(room *models.Room) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	
	if _, exists := rs.rooms[room.ID]; exists {
		return errors.New("room already exists")
	}
	
	rs.rooms[room.ID] = room
	return nil
}

func (rs *RoomService) RemoveRoom(id string) error {
	rs.mutex.Lock()
	defer rs.mutex.Unlock()
	
	if _, exists := rs.rooms[id]; !exists {
		return errors.New("room not found")
	}
	
	delete(rs.rooms, id)
	return nil
}

func (rs *RoomService) ListRooms() []*models.Room {
	rs.mutex.RLock()
	defer rs.mutex.RUnlock()
	
	rooms := make([]*models.Room, 0, len(rs.rooms))
	for _, room := range rs.rooms {
		rooms = append(rooms, room)
	}
	return rooms
}
