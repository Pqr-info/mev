// mev/core/src/zeta.rs

/// Zeta membrane core math: stress tensor S_ij, dynamic impedance Z_ij, preload strain P_ij.
#[derive(Debug, Clone, Copy, PartialEq)]
pub struct ZetaParams {
    pub alpha: f64,
    pub beta: f64,
    pub gamma: f64,
    pub z0: f64,
    pub theta_class: f64,
    pub k_preload: f64,
    pub s_max: f64,
    pub z_critical: f64,
    pub p_limit: f64,
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct CorridorState {
    pub s_ij: f64,
    pub z_ij: f64,
    pub p_ij: f64,
    pub rupture: bool,
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub enum CrisisType {
    DimensionalDrift,
    CrossMeshCollapse,
    EntropyWave,
}

#[derive(Debug, Clone, Copy, PartialEq)]
pub struct CrisisModifier {
    pub crisis_type: CrisisType,
    pub magnitude: f64,
}

impl CrisisModifier {
    pub fn apply(&self, base: &ZetaParams) -> ZetaParams {
        let mut p = *base;
        match self.crisis_type {
            CrisisType::DimensionalDrift => {
                p.theta_class /= 1.0 + 0.1 * self.magnitude;
                p.s_max /= 1.0 + 0.15 * self.magnitude;
                p.p_limit /= 1.0 + 0.1 * self.magnitude;
            }
            CrisisType::CrossMeshCollapse => {
                p.theta_class /= 1.0 + 0.15 * self.magnitude;
                p.s_max /= 1.0 + 0.1 * self.magnitude;
                p.z_critical /= 1.0 + 0.2 * self.magnitude;
            }
            CrisisType::EntropyWave => {
                p.theta_class /= 1.0 + 0.08 * self.magnitude;
                p.s_max /= 1.0 + 0.12 * self.magnitude;
                p.z_critical /= 1.0 + 0.15 * self.magnitude;
                p.p_limit /= 1.0 + 0.1 * self.magnitude;
            }
        }
        p
    }
}

/// Compute Euclidean norm of a 3D displacement vector.
#[inline]
pub fn norm3(d: (f64, f64, f64)) -> f64 {
    let (x, y, z) = d;
    (x * x + y * y + z * z).sqrt()
}

/// Compute stress tensor S_ij.
#[inline]
pub fn compute_stress(
    d_ij: (f64, f64, f64),
    delta_latency: f64,
    baseline_latency: f64,
    c_i: f64,
    c_j: f64,
    params: &ZetaParams,
) -> f64 {
    let disp_norm = norm3(d_ij);
    let displacement_term = params.alpha * disp_norm * disp_norm;
    let latency_term = params.beta * (delta_latency / baseline_latency);
    let compute_term = params.gamma * (c_i * c_j);
    displacement_term + latency_term + compute_term
}

/// Compute dynamic impedance Z_ij.
#[inline]
pub fn compute_impedance(s_ij: f64, params: &ZetaParams) -> f64 {
    params.z0 * (s_ij / params.theta_class).exp()
}

/// Compute preload strain P_ij.
#[inline]
pub fn compute_preload_strain(d_ij: (f64, f64, f64), params: &ZetaParams) -> f64 {
    params.k_preload * norm3(d_ij)
}

/// Evaluate full corridor state and rupture condition.
#[inline]
pub fn evaluate_corridor_state(
    d_ij: (f64, f64, f64),
    delta_latency: f64,
    baseline_latency: f64,
    c_i: f64,
    c_j: f64,
    base_params: &ZetaParams,
    crisis: Option<CrisisModifier>,
) -> CorridorState {
    let params = match crisis {
        Some(cm) => cm.apply(base_params),
        None => *base_params,
    };

    let s_ij = compute_stress(d_ij, delta_latency, baseline_latency, c_i, c_j, &params);
    let z_ij = compute_impedance(s_ij, &params);
    let p_ij = compute_preload_strain(d_ij, &params);

    let rupture = s_ij > params.s_max && z_ij > params.z_critical && p_ij > params.p_limit;

    CorridorState {
        s_ij,
        z_ij,
        p_ij,
        rupture,
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn default_params() -> ZetaParams {
        ZetaParams {
            alpha: 10.0,
            beta: 2.0,
            gamma: 5.0,
            z0: 50.0,
            theta_class: 8.0,
            k_preload: 15.0,
            s_max: 45.0,
            z_critical: 350.0,
            p_limit: 5.0,
        }
    }

    #[test]
    fn stable_conditions_no_rupture() {
        let params = default_params();
        let d_ij = (0.05, 0.02, -0.01);
        let delta_latency = 10.0;
        let baseline_latency = 10.0;
        let c_i = 0.3;
        let c_j = 0.35;

        let state = evaluate_corridor_state(d_ij, delta_latency, baseline_latency, c_i, c_j, &params, None);

        assert!(state.s_ij > 0.0);
        assert!(state.z_ij > 0.0);
        assert!(state.p_ij < params.p_limit);
        assert!(!state.rupture);
    }

    #[test]
    fn consensus_spike_high_stress_but_preload_holds() {
        let mut params = default_params();
        params.s_max = 50.0;
        params.z_critical = 500.0;
        params.p_limit = 5.0;

        let d_ij = (0.05, 0.02, -0.01); // minimal displacement
        let delta_latency = 95.0;
        let baseline_latency = 10.0;
        let c_i = 0.90;
        let c_j = 0.95;

        let state = evaluate_corridor_state(d_ij, delta_latency, baseline_latency, c_i, c_j, &params, None);

        assert!(state.s_ij > 0.0);
        assert!(state.z_ij > 0.0);
        assert!(state.s_ij > params.s_max || state.z_ij > params.z_critical);
        assert!(state.p_ij < params.p_limit);
        assert!(!state.rupture);
    }

    #[test]
    fn catastrophic_flux_storm_triggers_rupture() {
        let params = default_params();

        let d_ij = (0.6, -0.4, 0.35); // large displacement
        let delta_latency = 180.0;
        let baseline_latency = 10.0;
        let c_i = 0.98;
        let c_j = 0.99;

        let state = evaluate_corridor_state(d_ij, delta_latency, baseline_latency, c_i, c_j, &params, None);

        assert!(state.s_ij > params.s_max);
        assert!(state.z_ij > params.z_critical);
        assert!(state.p_ij > params.p_limit);
        assert!(state.rupture);
    }

    #[test]
    fn norm3_is_correct() {
        let v = (0.3_f64, 0.4_f64, 0.0_f64);
        let n = norm3(v);
        assert!((n - 0.5).abs() < 1e-9);
    }

    #[test]
    fn dimensional_drift_tightens_limits() {
        let b = default_params();
        let cm = CrisisModifier {
            crisis_type: CrisisType::DimensionalDrift,
            magnitude: 1.5,
        };
        let p = cm.apply(&b);
        assert!(p.theta_class < b.theta_class);
        assert!(p.s_max < b.s_max);
        assert!(p.p_limit < b.p_limit);
    }

    #[test]
    fn cross_mesh_collapse_compresses_tolerance() {
        let b = default_params();
        let cm = CrisisModifier {
            crisis_type: CrisisType::CrossMeshCollapse,
            magnitude: 1.0,
        };
        let p = cm.apply(&b);
        assert!(p.theta_class < b.theta_class);
        assert!(p.s_max < b.s_max);
        assert!(p.z_critical < b.z_critical);
    }

    #[test]
    fn entropy_wave_lowers_all_limits() {
        let b = default_params();
        let cm = CrisisModifier {
            crisis_type: CrisisType::EntropyWave,
            magnitude: 2.0,
        };
        let p = cm.apply(&b);
        assert!(p.theta_class < b.theta_class);
        assert!(p.s_max < b.s_max);
        assert!(p.p_limit < b.p_limit);
        assert!(p.z_critical < b.z_critical);
    }
}
