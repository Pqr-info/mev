/**
 * Dolphin Location Proximity, Bluetooth & Wi-Fi Direct Mesh Discovery Engine
 */

class LocationEngine {
  constructor() {
    this.currentLocation = {
      latitude: 37.7749,
      longitude: -122.4194,
      accuracy: 10,
      timestamp: Date.now()
    };
    this.listeners = [];
    this.watchId = null;
    this.isBleScanning = false;
    this.discoveredBleDevices = [];

    this.initGeolocation();
  }

  initGeolocation() {
    if ('geolocation' in navigator) {
      this.watchId = navigator.geolocation.watchPosition(
        (pos) => {
          this.currentLocation = {
            latitude: pos.coords.latitude,
            longitude: pos.coords.longitude,
            accuracy: pos.coords.accuracy,
            timestamp: pos.timestamp
          };
          this.notify();
        },
        (err) => {
          console.warn('[Proximity Engine] Geolocation fallback used:', err.message);
        },
        { enableHighAccuracy: true, timeout: 10000, maximumAge: 1000 }
      );
    }
  }

  calculateDistance(lat1, lon1, lat2, lon2) {
    const R = 6371e3;
    const rad = Math.PI / 180;
    const dLat = (lat2 - lat1) * rad;
    const dLon = (lon2 - lon1) * rad;
    const a =
      Math.sin(dLat / 2) * Math.sin(dLat / 2) +
      Math.cos(lat1 * rad) * Math.cos(lat2 * rad) * Math.sin(dLon / 2) * Math.sin(dLon / 2);
    const c = 2 * Math.atan2(Math.sqrt(a), Math.sqrt(1 - a));
    return Math.round(R * c);
  }

  getNearbyPeers() {
    const { latitude, longitude } = this.currentLocation;

    return [
      {
        id: 'peer_1',
        name: 'Sarah Jenkins',
        username: '@sjenkins',
        avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=250&q=80',
        lat: latitude + 0.0012,
        lng: longitude + 0.0015,
        distanceMeters: this.calculateDistance(latitude, longitude, latitude + 0.0012, longitude + 0.0015),
        status: 'Exploring Solar Portal anomaly ☀️',
        angle: 45,
        protocol: 'Wi-Fi Direct P2P'
      },
      {
        id: 'peer_2',
        name: 'Alex Rivera',
        username: '@arivera',
        avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=250&q=80',
        lat: latitude - 0.0021,
        lng: longitude + 0.0018,
        distanceMeters: this.calculateDistance(latitude, longitude, latitude - 0.0021, longitude + 0.0018),
        status: 'Harmonic frequency 528Hz aligned 🎶',
        angle: 135,
        protocol: 'Bluetooth LE Mesh'
      },
      {
        id: 'peer_3',
        name: 'Elena Rostova',
        username: '@elena_r',
        avatar: 'https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&w=250&q=80',
        lat: latitude - 0.0035,
        lng: longitude - 0.0028,
        distanceMeters: this.calculateDistance(latitude, longitude, latitude - 0.0035, longitude - 0.0028),
        status: 'Ocean Beach Cafe ☕',
        angle: 225,
        protocol: 'Wi-Fi P2P'
      },
      {
        id: 'peer_4',
        name: 'Marcus Chen',
        username: '@mchen',
        avatar: 'https://images.unsplash.com/photo-1500648767791-00dcc994a43e?auto=format&fit=crop&w=250&q=80',
        lat: latitude + 0.0042,
        lng: longitude - 0.0015,
        distanceMeters: this.calculateDistance(latitude, longitude, latitude + 0.0042, longitude - 0.0015),
        status: 'Sovereign-27 Proximity Node Active ⚡',
        angle: 315,
        protocol: 'Bluetooth 5.3 Mesh'
      }
    ];
  }

  async scanBluetoothWifiNeighbors() {
    this.isBleScanning = true;
    let bleResult = null;

    if (navigator.bluetooth) {
      try {
        const device = await navigator.bluetooth.requestDevice({
          acceptAllDevices: true,
          optionalServices: ['battery_service', 'device_information']
        });
        bleResult = {
          name: device.name || 'Sovereign BLE Mesh Node',
          id: device.id,
          connected: device.gatt ? device.gatt.connected : false,
          protocol: 'Web Bluetooth LE'
        };
      } catch (err) {
        console.warn('Web Bluetooth user prompt cancelled or unavailable:', err.message);
      }
    }

    // Simulated BLE / Wi-Fi Direct Peer Beacon Discovery Scan
    const simulatedNeighbors = [
      {
        id: 'ble_node_01',
        name: 'BLE Beacon Node [zeta-01]',
        mac: '7A:3F:89:D2:11:04',
        rssi: -58,
        distanceMeters: 3.2,
        protocol: 'BLE 5.2 Mesh',
        services: ['Sovereign-27-GMI', 'P2P-Sync']
      },
      {
        id: 'wifi_direct_02',
        name: 'Wi-Fi Direct Peer [ted-mesh]',
        mac: 'BE:EF:46:22:19:74',
        rssi: -42,
        distanceMeters: 8.5,
        protocol: 'Wi-Fi Direct P2P (5GHz)',
        services: ['NBEP-Substrate', 'rqlite-Bridge']
      },
      {
        id: 'ble_node_03',
        name: 'Proximity Node [max-relay]',
        mac: 'C0:FF:EE:27:00:49',
        rssi: -65,
        distanceMeters: 14.1,
        protocol: 'BLE Long-Range Coded',
        services: ['PQLite-WAL', 'Shared-Brain']
      }
    ];

    if (bleResult) {
      simulatedNeighbors.unshift({
        id: `ble_real_${Date.now()}`,
        name: bleResult.name,
        mac: bleResult.id.substring(0, 17),
        rssi: -38,
        distanceMeters: 1.5,
        protocol: bleResult.protocol,
        services: ['Web-Bluetooth-Active']
      });
    }

    this.discoveredBleDevices = simulatedNeighbors;
    this.isBleScanning = false;
    return simulatedNeighbors;
  }

  subscribe(callback) {
    this.listeners.push(callback);
    return () => {
      this.listeners = this.listeners.filter(cb => cb !== callback);
    };
  }

  notify() {
    this.listeners.forEach(cb => cb(this.currentLocation));
  }
}

export const locationEngine = new LocationEngine();
