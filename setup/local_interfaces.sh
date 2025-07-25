#!/bin/bash

# add_loopback_ip.sh
# Adds multiple IP addresses to loopback interface

IPS=("127.0.0.101" "127.0.0.102" "127.0.0.103")
INTERFACE="lo"

echo "Adding IP addresses to loopback interface..."

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "This script must be run as root (use sudo)"
    exit 1
fi

# Add each IP address to loopback interface
for IP in "${IPS[@]}"; do
    echo "Adding $IP to $INTERFACE..."
    ip addr add $IP/32 dev $INTERFACE
    
    # Check if the command was successful
    if [ $? -eq 0 ]; then
        echo "Successfully added $IP to $INTERFACE"
    else
        echo "Failed to add $IP to $INTERFACE"
        exit 1
    fi
done

echo "Verifying all IP addresses..."
for IP in "${IPS[@]}"; do
    ip addr show $INTERFACE | grep $IP
done
