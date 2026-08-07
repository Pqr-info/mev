/**
 * Human Design Bodygraph & Relational Composite Engine
 * Computes individual 9 Energy Center definitions, gates, channels, and multi-person composite dynamics.
 */

export const ENERGY_CENTERS = {
  HEAD: { id: 'HEAD', name: 'Head Center', type: 'Inspiration', shape: 'triangle-up' },
  AJNA: { id: 'AJNA', name: 'Ajna Center', type: 'Conceptualization', shape: 'triangle-down' },
  THROAT: { id: 'THROAT', name: 'Throat Center', type: 'Manifestation', shape: 'square' },
  G_CENTER: { id: 'G_CENTER', name: 'G-Center', type: 'Identity & Direction', shape: 'diamond' },
  HEART: { id: 'HEART', name: 'Heart / Ego Center', type: 'Willpower', shape: 'triangle-right' },
  SOLAR_PLEXUS: { id: 'SOLAR_PLEXUS', name: 'Solar Plexus', type: 'Emotional Awareness', shape: 'triangle-right' },
  SACRAL: { id: 'SACRAL', name: 'Sacral Center', type: 'Life Force Energy', shape: 'square' },
  SPLEEN: { id: 'SPLEEN', name: 'Spleen Center', type: 'Intuition & Health', shape: 'triangle-left' },
  ROOT: { id: 'ROOT', name: 'Root Center', type: 'Adrenaline Pressure', shape: 'square' }
};

export const HUMAN_DESIGN_PROFILES = [
  {
    userId: 'u101',
    name: 'Antigravity Dev (You)',
    avatar: 'https://images.unsplash.com/photo-1534528741775-53994a69daeb?auto=format&fit=crop&w=250&q=80',
    type: 'Manifesting Generator',
    profile: '1/3 Investigator/Martyr',
    authority: 'Sacral Authority',
    definition: 'Split Definition',
    activeGates: [34, 20, 64, 47, 1, 8, 59, 6],
    definedCenters: ['SACRAL', 'THROAT', 'HEAD', 'AJNA', 'G_CENTER']
  },
  {
    userId: 'peer_1',
    name: 'Sarah Jenkins',
    avatar: 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?auto=format&fit=crop&w=250&q=80',
    type: 'Projector',
    profile: '2/4 Hermit/Opportunist',
    authority: 'Emotional Authority',
    definition: 'Single Definition',
    activeGates: [43, 23, 28, 38, 39, 55, 37, 40],
    definedCenters: ['AJNA', 'THROAT', 'SPLEEN', 'SOLAR_PLEXUS', 'HEART']
  },
  {
    userId: 'peer_2',
    name: 'Alex Rivera',
    avatar: 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?auto=format&fit=crop&w=250&q=80',
    type: 'Manifestor',
    profile: '4/6 Opportunist/Role Model',
    authority: 'Splenic Authority',
    definition: 'Single Definition',
    activeGates: [21, 45, 10, 20, 28, 38],
    definedCenters: ['HEART', 'THROAT', 'G_CENTER', 'SPLEEN', 'ROOT']
  },
  {
    userId: 'peer_3',
    name: 'Elena Rostova',
    avatar: 'https://images.unsplash.com/photo-1517841905240-472988babdf9?auto=format&fit=crop&w=250&q=80',
    type: 'Generator',
    profile: '5/1 Heretic/Investigator',
    authority: 'Sacral Authority',
    definition: 'Single Definition',
    activeGates: [59, 6, 34, 20, 64, 47],
    definedCenters: ['SACRAL', 'SOLAR_PLEXUS', 'THROAT', 'HEAD', 'AJNA']
  }
];

export const CHANNELS_MAP = [
  { id: '34-20', name: 'Channel of Charisma', from: 'SACRAL', to: 'THROAT', gates: [34, 20] },
  { id: '64-47', name: 'Channel of Abstraction', from: 'HEAD', to: 'AJNA', gates: [64, 47] },
  { id: '43-23', name: 'Channel of Structuring', from: 'AJNA', to: 'THROAT', gates: [43, 23] },
  { id: '1-8', name: 'Channel of Inspiration', from: 'G_CENTER', to: 'THROAT', gates: [1, 8] },
  { id: '10-20', name: 'Channel of Awakening', from: 'G_CENTER', to: 'THROAT', gates: [10, 20] },
  { id: '59-6', name: 'Channel of Mating & Friction', from: 'SACRAL', to: 'SOLAR_PLEXUS', gates: [59, 6] },
  { id: '28-38', name: 'Channel of Struggle', from: 'SPLEEN', to: 'ROOT', gates: [28, 38] },
  { id: '39-55', name: 'Channel of Emoting', from: 'ROOT', to: 'SOLAR_PLEXUS', gates: [39, 55] },
  { id: '21-45', name: 'Channel of Money', from: 'HEART', to: 'THROAT', gates: [21, 45] },
  { id: '37-40', name: 'Channel of Community', from: 'SOLAR_PLEXUS', to: 'HEART', gates: [37, 40] }
];

class HumanDesignEngine {
  constructor() {
    this.profiles = [...HUMAN_DESIGN_PROFILES];
  }

  getProfileByUserId(userId) {
    return this.profiles.find(p => p.userId === userId) || this.profiles[0];
  }

  /**
   * Calculates Relational Composite for 2 Participants (Couples)
   */
  calculateCoupleComposite(userAId, userBId) {
    const personA = this.getProfileByUserId(userAId);
    const personB = this.getProfileByUserId(userBId);

    const combinedGates = Array.from(new Set([...personA.activeGates, ...personB.activeGates]));
    const combinedCenters = Array.from(new Set([...personA.definedCenters, ...personB.definedCenters]));

    // Identify Electromagnetic Channels (where Gate 1 is in Person A and Gate 2 is in Person B)
    const electromagneticChannels = [];
    const dominantChannels = [];
    const compromiseChannels = [];

    CHANNELS_MAP.forEach(channel => {
      const [g1, g2] = channel.gates;
      const hasA1 = personA.activeGates.includes(g1);
      const hasA2 = personA.activeGates.includes(g2);
      const hasB1 = personB.activeGates.includes(g1);
      const hasB2 = personB.activeGates.includes(g2);

      const aComplete = hasA1 && hasA2;
      const bComplete = hasB1 && hasB2;

      if ((hasA1 && hasB2 && !hasA2 && !hasB1) || (hasA2 && hasB1 && !hasA1 && !hasB2)) {
        electromagneticChannels.push({ ...channel, type: 'ELECTROMAGNETIC', label: 'Electromagnetic Connection 🔥' });
      } else if (aComplete && !hasB1 && !hasB2) {
        dominantChannels.push({ ...channel, type: 'DOMINANCE_A', label: `${personA.name} Dominance` });
      } else if (bComplete && !hasA1 && !hasA2) {
        dominantChannels.push({ ...channel, type: 'DOMINANCE_B', label: `${personB.name} Dominance` });
      } else if ((aComplete && (hasB1 || hasB2)) || (bComplete && (hasA1 || hasA2))) {
        compromiseChannels.push({ ...channel, type: 'COMPROMISE', label: 'Compromise Channel ⚖️' });
      }
    });

    const definedCenterCount = combinedCenters.length;
    let relationshipType = '9-0 Nowhere to Go';
    if (definedCenterCount === 9) relationshipType = '9-0 Full Harmony';
    else if (definedCenterCount === 8) relationshipType = '8-1 Have Fun';
    else if (definedCenterCount === 7) relationshipType = '7-2 Work to Do';
    else if (definedCenterCount === 6) relationshipType = '6-3 Free Spirit';

    return {
      personA,
      personB,
      combinedGates,
      combinedCenters,
      electromagneticChannels,
      dominantChannels,
      compromiseChannels,
      relationshipType,
      synergyScore: Math.min(99, 65 + (electromagneticChannels.length * 8))
    };
  }

  /**
   * Calculates Group Penta Dynamics for 3 to 9 Participants
   */
  calculateGroupPenta(userIds) {
    const selectedProfiles = userIds.map(id => this.getProfileByUserId(id));
    const allGates = Array.from(new Set(selectedProfiles.flatMap(p => p.activeGates)));
    const allCenters = Array.from(new Set(selectedProfiles.flatMap(p => p.definedCenters)));

    // Group Penta Synergy Calculation
    const typeCounts = {};
    selectedProfiles.forEach(p => {
      typeCounts[p.type] = (typeCounts[p.type] || 0) + 1;
    });

    const pentaDefinedCentersCount = allCenters.length;
    const synergyScore = Math.min(100, Math.round((pentaDefinedCentersCount / 9) * 100));

    return {
      membersCount: selectedProfiles.length,
      profiles: selectedProfiles,
      allGates,
      allCenters,
      pentaDefinedCentersCount,
      typeCounts,
      synergyScore,
      pentaStatus: pentaDefinedCentersCount >= 7 ? 'High Alignment Synergy ⚡' : 'Moderate Alignment'
    };
  }
}

export const humanDesignEngine = new HumanDesignEngine();
