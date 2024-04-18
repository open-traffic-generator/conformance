#!/bin/bash

#change this to your UHD IP/Hostname
UHD_IP=10.36.87.166

URL="https://${UHD_IP}/port/api/v1/config"
CONTENT_TYPE="application/json"
DATA='{
  "ports": [
    {"port_number": 1,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 2,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 3,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 4,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 5,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 6,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 7,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 8,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 9,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 10,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 11,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 12,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 13,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 14,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 15,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {"port_number": 16,"enabled": true,"settings": {"speed": "speed_400_gbps","fec_mode": "reed_solomon","negotiation": {"choice": "manual"}}},
    {
      "port_number": 32,
      "enabled": true,
      "settings": {
        "speed": "speed_10_gbps",
        "negotiation": {
          "choice": "manual"
        }
      }       
    }  
  ]
}'

# Send the HTTP request using curl
curl -fLvk "$URL" -X PUT -H "Content-Type: $CONTENT_TYPE" -d "$DATA"
echo "Speed set successfully"
