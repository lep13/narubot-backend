package services

import (
    "bytes"
    "encoding/json"
    "net/http"
)

// Structure for Webex message data
type WebexMessage struct {
    RoomId string `json:"roomId"`
    Text   string `json:"text"`
}

// Function to send a message back to Webex
func SendMessageToWebex(roomId string, message string, accessToken string) error {
    client := &http.Client{}
    messageData := WebexMessage{
        RoomId: roomId,
        Text:   message,
    }

    jsonData, _ := json.Marshal(messageData)

    req, err := http.NewRequest("POST", "https://webexapis.com/v1/messages", bytes.NewBuffer(jsonData))
    if err != nil {
        return err
    }

    req.Header.Set("Content-Type", "application/json")
    req.Header.Set("Authorization", "Bearer "+accessToken)

    _, err = client.Do(req)
    return err
}
