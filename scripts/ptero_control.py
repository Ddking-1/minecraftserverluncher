#!/usr/bin/env python3
"""
Pterodactyl Control Script for Minecraft Server Management.
This script provides control over Pterodactyl servers via the Application API,
including server listing, status checking, console log reading, and start/stop/restart operations.
"""

import requests
import os
from dotenv import load_dotenv

# Load environment variables from .env file
load_dotenv()

# Configuration
PTERO_URL = os.getenv("PTERO_URL")
PTERO_APP_API_KEY = os.getenv("PTERO_APP_API_KEY")

if not PTERO_URL or not PTERO_APP_API_KEY:
    raise ValueError("Missing required environment variables: PTERO_URL, PTERO_APP_API_KEY")

# Headers for API requests
HEADERS = {
    "Authorization": f"Bearer {PTERO_APP_API_KEY}",
    "Content-Type": "application/json"
}

def list_servers():
    """List all servers."""
    url = f"{PTERO_URL}/api/application/servers?per_page=100"
    response = requests.get(url, headers=HEADERS)
    return response.json()

def get_server_status(server_id):
    """Get the status of a specific server by ID."""
    url = f"{PTERO_URL}/api/application/servers/{server_id}"
    response = requests.get(url, headers=HEADERS)
    if response.status_code == 200:
        data = response.json()
        return {
            "status": data['attributes']['current_state'],
            "public_id": data['attributes']['identifier']
        }
    else:
        raise Exception(f"Failed to get server status: {response.status_code}")

def read_console_logs(server_id, lines=100):
    """Read the console logs for a specific server."""
    url = f"{PTERO_URL}/api/application/servers/{server_id}/logs"
    params = {"lines": lines}
    response = requests.get(url, headers=HEADERS, params=params)
    if response.status_code == 200:
        return response.text
    else:
        raise Exception(f"Failed to read console logs: {response.status_code}")

def start_server(server_id):
    """Start a specific server."""
    url = f"{PTERO_URL}/api/application/servers/{server_id}/power"
    payload = {"action": "start"}
    response = requests.post(url, headers=HEADERS, json=payload)
    return response.json()

def stop_server(server_id):
    """Stop a specific server."""
    url = f"{PTERO_URL}/api/application/servers/{server_id}/power"
    payload = {"action": "stop"}
    response = requests.post(url, headers=HEADERS, json=payload)
    return response.json()

def restart_server(server_id):
    """Restart a specific server."""
    url = f"{PTERO_URL}/api/application/servers/{server_id}/power"
    payload = {"action": "restart"}
    response = requests.post(url, headers=HEADERS, json=payload)
    return response.json()

def backup_server(server_id):
    """Trigger a backup for the specified server."""
    url = f"{PTERO_URL}/api/application/servers/{server_id}/backups"
    payload = {"name": "backup", "description": "Automated backup"}
    response = requests.post(url, headers=HEADERS, json=payload)
    return response.json()

if __name__ == "__main__":
    # Example usage (uncomment to test):
    # servers = list_servers()
    # print(servers)

    pass