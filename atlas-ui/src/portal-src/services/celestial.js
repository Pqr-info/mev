/**
 * 5D Celestial Mesh Engine
 * Manages location-based celestial objects, harmonic frequencies, and alignment mechanics.
 */

import { locationEngine } from './location.js';

export const CELESTIAL_TYPES = {
  SOLAR_FLARE: {
    type: 'SOLAR_FLARE',
    name: 'Solar Flare Beacon',
    icon: 'fa-sun',
    color: '#fb923c',
    defaultFreq: 432,
    rarity: 'Common',
    energy: 150
  },
  ASTRAL_BEACON: {
    type: 'ASTRAL_BEACON',
    name: 'Astral Crystal Node',
    icon: 'fa-gem',
    color: '#06b6d4',
    defaultFreq: 528,
    rarity: 'Rare',
    energy: 350
  },
  PULSAR_PORTAL: {
    type: 'PULSAR_PORTAL',
    name: 'Pulsar Star Gateway',
    icon: 'fa-atom',
    color: '#a855f7',
    defaultFreq: 639,
    rarity: 'Epic',
    energy: 750
  },
  COSMIC_CORE: {
    type: 'COSMIC_CORE',
    name: 'Lyran Quantum Core',
    icon: 'fa-bahai',
    color: '#f43f5e',
    defaultFreq: 963,
    rarity: 'Legendary',
    energy: 1500
  }
};

class CelestialEngine {
  constructor() {
    this.userQuantumEnergy = 1450;
    this.alignedObjects = [];
    this.celestialObjects = [];
    this.initDefaultObjects();
  }

  initDefaultObjects() {
    const loc = locationEngine.currentLocation;

    this.celestialObjects = [
      {
        id: 'c_obj_1',
        ...CELESTIAL_TYPES.SOLAR_FLARE,
        lat: loc.latitude + 0.0008,
        lng: loc.longitude + 0.0009,
        distanceMeters: locationEngine.calculateDistance(loc.latitude, loc.longitude, loc.latitude + 0.0008, loc.longitude + 0.0009),
        angle: 30,
        expiresIn: '45m',
        targetFreq: 432
      },
      {
        id: 'c_obj_2',
        ...CELESTIAL_TYPES.ASTRAL_BEACON,
        lat: loc.latitude - 0.0014,
        lng: loc.longitude + 0.0012,
        distanceMeters: locationEngine.calculateDistance(loc.latitude, loc.longitude, loc.latitude - 0.0014, loc.longitude + 0.0012),
        angle: 120,
        expiresIn: '1h 20m',
        targetFreq: 528
      },
      {
        id: 'c_obj_3',
        ...CELESTIAL_TYPES.PULSAR_PORTAL,
        lat: loc.latitude + 0.0022,
        lng: loc.longitude - 0.0018,
        distanceMeters: locationEngine.calculateDistance(loc.latitude, loc.longitude, loc.latitude + 0.0022, loc.longitude - 0.0018),
        angle: 300,
        expiresIn: '2h 10m',
        targetFreq: 639
      }
    ];
  }

  getCelestialObjects(radiusMeters = 5000) {
    const loc = locationEngine.currentLocation;
    return this.celestialObjects
      .map(obj => {
        const dist = locationEngine.calculateDistance(loc.latitude, loc.longitude, obj.lat, obj.lng);
        return { ...obj, distanceMeters: dist };
      })
      .filter(obj => obj.distanceMeters <= radiusMeters);
  }

  spawnAnomaly(typeKey = 'SOLAR_FLARE', customName = null) {
    const loc = locationEngine.currentLocation;
    const baseType = CELESTIAL_TYPES[typeKey] || CELESTIAL_TYPES.SOLAR_FLARE;

    // Random small offset around current location
    const latOffset = (Math.random() - 0.5) * 0.004;
    const lngOffset = (Math.random() - 0.5) * 0.004;
    const angle = Math.floor(Math.random() * 360);

    const newObj = {
      id: `c_obj_${Date.now()}`,
      ...baseType,
      name: customName || baseType.name,
      lat: loc.latitude + latOffset,
      lng: loc.longitude + lngOffset,
      distanceMeters: locationEngine.calculateDistance(loc.latitude, loc.longitude, loc.latitude + latOffset, loc.longitude + lngOffset),
      angle,
      expiresIn: '2h 00m',
      targetFreq: baseType.defaultFreq
    };

    this.celestialObjects.unshift(newObj);
    return newObj;
  }

  alignCelestialObject(objectId, tunedFreq) {
    const objIndex = this.celestialObjects.findIndex(o => o.id === objectId);
    if (objIndex === -1) return { success: false, message: 'Object expired or out of range' };

    const obj = this.celestialObjects[objIndex];
    const freqDiff = Math.abs(obj.targetFreq - tunedFreq);

    if (freqDiff <= 10) {
      // Successful Harmonic Alignment
      this.userQuantumEnergy += obj.energy;
      this.alignedObjects.push({ ...obj, alignedAt: new Date().toLocaleTimeString() });
      this.celestialObjects.splice(objIndex, 1);

      return {
        success: true,
        rewardEnergy: obj.energy,
        object: obj,
        message: `Harmonic alignment complete! +${obj.energy} Quantum Energy added.`
      };
    }

    return {
      success: false,
      message: `Frequency mismatch (Diff: ${freqDiff}Hz). Adjust slider closer to ${obj.targetFreq}Hz.`
    };
  }
}

export const celestialEngine = new CelestialEngine();
