# The Science of Rock Temperature for Climbing

## Overview

This document explains the physics and reasoning behind the woulder.com rock temperature calculator. Every formula, constant, and threshold has a basis in either heat-transfer physics or climber field observations. Sources are listed at the end.

For implementation details (code structure, database migrations, integration), see the companion `rock-temperature.md` design document.

---

## Table of Contents

1. [Why Rock Temperature Isn't Air Temperature](#1-why-rock-temperature-isnt-air-temperature)
2. [The Energy Balance](#2-the-energy-balance)
3. [Solar Heating](#3-solar-heating)
4. [Sky Radiation and Nighttime Cooling](#4-sky-radiation-and-nighttime-cooling)
5. [Wind and Convective Cooling](#5-wind-and-convective-cooling)
6. [Thermal Mass and Lag](#6-thermal-mass-and-lag)
7. [Rock Types](#7-rock-types)
8. [Friction Quality Thresholds](#8-friction-quality-thresholds)
9. [Condensation](#9-condensation)
10. [Worked Example](#10-worked-example)
11. [Confidence and Limitations](#11-confidence-and-limitations)
12. [Sources](#12-sources)

---

## 1. Why Rock Temperature Isn't Air Temperature

Rock temperature can differ from air temperature by 30°F or more. Rock and air heat by different mechanisms: air mostly through convection and mixing, rock through absorbing sunlight. On a clear day with no wind, dark rock in direct sun can sit 30–40°F above air temperature (Jin & Dickinson 2010). At night under a clear sky, that same rock can drop 10–15°F *below* air temperature.

A weather forecast tells you what air is doing. Predicting climbing conditions requires actually modeling the rock.

> **Analogy:** A black car and a white car parked in the same air heat very differently in the sun. The air is one thing; the surface is another.

---

## 2. The Energy Balance

A rock surface is constantly gaining heat from some sources and losing it to others. When they balance, the rock holds steady. That steady temperature is what we're predicting.

Four energy flows at a rock surface:

| Flow | Direction | Driver |
|------|-----------|--------|
| **Solar absorption** | In | Sunlight |
| **Sky radiation** | Out (mostly) | Rock emits IR to cold sky |
| **Convection** | In or out | Wind moves heat between surface and air |
| **Conduction into rock** | In or out | Heat diffuses into the interior |

At equilibrium, conduction into deep rock averages to zero, and:

```
Solar in = Sky radiation out + Convection out
```

Solving for surface temperature, the linearized form is:

$$T_{eq} = T_{air} + \frac{\alpha \cdot I_{face} - h_{rad} \cdot (T_{air} - T_{sky})}{h_{conv} + h_{rad}}$$

This is the core equation of the system. Every other piece of the model exists to compute one of its terms:

- **α (absorptivity):** fraction of sunlight the rock absorbs vs reflects (0.55–0.95 depending on color)
- **I_face:** effective sunlight hitting this specific face (W/m²) — accounts for sun position, clouds, trees, aspect, dip
- **h_conv:** convective heat transfer coefficient — bigger when windier
- **h_rad ≈ 5.5 W/m²·K:** linearized radiation coefficient
- **T_sky:** effective sky temperature — clear sky ~20°F below air, overcast ≈ air temperature

**Why linearize?** True radiation goes as T⁴ (Stefan-Boltzmann). Linearizing around air temperature loses <5% accuracy over climbing-relevant temps (0°F to 130°F), far less than uncertainty in other inputs (Bergman et al., *Fundamentals of Heat and Mass Transfer*).

---

## 3. Solar Heating

Three things determine solar input to a rock face: how much sunlight is in the sky, how well the face is aimed at it, and what fraction the rock absorbs.

### 3.1 Direct + Diffuse

Sunlight has two components:

- **Direct:** the beam from the sun (casts a shadow). Up to ~900 W/m² horizontal at noon on a clear day.
- **Diffuse:** scattered light from the rest of the sky. ~15% of total on clear days, 100% under full overcast.

Open-Meteo provides both (`direct_radiation`, `diffuse_radiation`, both on a horizontal surface). They must be handled separately: a vertical wall sees direct only when sun azimuth aligns with its aspect, but sees diffuse continuously.

### 3.2 Lambert's Cosine Law

A surface catches the perpendicular component of incoming light. Tilt the surface and the same beam is spread over more area (Lambert 1760):

$$I_{on\_face} = I_{DNI} \cdot \cos(\theta)$$

where θ is the angle between the sun's direction and the face's normal.

For a face with aspect (compass direction) and dip (tilt from horizontal), with the sun at elevation `elev` and azimuth `az`:

$$\cos(\theta) = \sin(elev) \cdot \cos(dip) + \cos(elev) \cdot \sin(dip) \cdot \cos(az - aspect)$$

Special cases:

- **Horizontal slab (dip = 0):** cos(θ) = sin(elev). Pure altitude dependence.
- **Vertical wall (dip = 90°):** cos(θ) = cos(elev) · cos(az − aspect). Best when sun is low *and* aligned with the wall's compass direction.

This explains why **south-facing walls cook in winter**: low sun + perpendicular alignment with a vertical wall = maximum cosine factor. Flat slabs do poorly in winter because sin(low elev) is small.

> **Analogy:** Sunlight is a firehose. You catch the most water when the nozzle aims straight at you. Turn sideways and most of it misses.

### 3.3 The Horizontal vs Normal Conversion

The most common solar-tilt implementation mistake: Open-Meteo's `direct_radiation` is **measured on a horizontal surface**, not the beam strength itself. To apply the cosine rule to a tilted face, convert to direct-normal irradiance (DNI) first:

$$I_{DNI} = \frac{I_{direct\_horizontal}}{\sin(elev)}$$

Skipping this step double-applies the elevation correction — you'd get correct numbers for horizontal slabs and wrong answers for everything else.

Guard against very low sun: when elev < 5°, set DNI = 0 (avoids dividing by tiny numbers; atmospheric scattering kills the direct beam at those angles anyway).

### 3.4 Tree Shade and Sky View Factor

**Tree coverage** reduces direct light by ~75% per unit canopy and diffuse by ~50% (forest microclimate research; Federer 1971; Hardy et al. 2004). Diffuse isn't reduced as much because the canopy itself becomes a secondary diffuse-light source.

**Sky view factor** — the fraction of sky a face can "see":

$$F_{sky} = \frac{1 + \cos(dip)}{2}$$

Horizontal slab: F_sky = 1.0 (sees full hemisphere). Vertical wall: F_sky = 0.5 (sees half — the other half is ground or wall behind). This affects both diffuse light and sky radiation cooling.

### 3.5 Combined Irradiance

$$I_{face} = I_{DNI} \cdot \cos(\theta) \cdot (1 - tree \cdot 0.75) + I_{diffuse} \cdot F_{sky} \cdot (1 - tree \cdot 0.5)$$

Multiplying by α gives the heat input per square meter.

### 3.6 Absorptivity by Rock Type

| Rock | α | Notes |
|------|---|-------|
| Granite | 0.65–0.75 | Light gray |
| Quartzite | 0.65–0.75 | Light, glassy |
| Limestone | 0.55–0.65 | Lightest of common climbing rocks |
| Sandstone | 0.70–0.80 | Varies with color; varnish darkens it |
| Basalt / Gabbro | 0.85–0.95 | Very dark, nearly black |

Sources: Clark 1966 (*Handbook of Physical Constants*), ASHRAE Handbook surface property tables.

A basalt boulder (α ≈ 0.90) absorbs ~30% more sunlight than a granite boulder (α ≈ 0.70). Under 800 W/m² of effective sun, that's an extra 160 W/m² — over a couple hours, the difference between "tolerable" and "unsendable."

---

## 4. Sky Radiation and Nighttime Cooling

Everything with a temperature radiates electromagnetic energy. The rock radiates infrared. So does the sky — but at very different effective temperatures.

**Clear sky behaves like a very cold object**, often around 0°F on a summer night. You're effectively looking through the atmosphere into the cold of space. Some water vapor and CO₂ re-emit warmth back, but on a dry clear night, the sky is a heat sink at near-space temperatures (Berdahl & Fromberg 1982).

**Cloudy sky behaves close to air temperature.** Clouds are at roughly air temperature and emit IR back at the ground.

This is why:

- Frost forms on clear nights, not cloudy nights, even when air doesn't quite freeze.
- Desert day-night swings are enormous (dry air → clear sky → strong radiative cooling).
- Cloudy nights stay warm (clouds act as a thermal blanket).

In the model:

$$T_{sky} \approx T_{air} - 20°F \quad \text{(clear)}$$
$$T_{sky} \approx T_{air} \quad \text{(overcast)}$$

Linearly interpolated by cloud cover. The 20°F clear-sky offset is typical for mid-latitudes; deep deserts hit 30–40°F.

**Why this matters for climbing:** A south-facing granite wall might absorb 100 W/m² of diffuse sky during a cloudless day but radiate ~150 W/m² of net IR at night to that −20°F effective sky. By 4–5 AM the granite can sit 5–15°F *below* air temperature, even after baking 20°F above air the previous afternoon. That's the predawn send window every desert climber knows.

Without this term, the model predicts T_rock = T_air at night and the predawn window disappears.

> **Analogy:** On a clear night the windshield of your car frosts before anything else — it's angled at the open sky, radiating warmth into space. Under a tree or carport, no frost. The windshield sees a warm "sky" instead.

---

## 5. Wind and Convective Cooling

Wind moves heat between rock and air. Whichever is warmer loses to whichever is cooler. **Wind pushes rock toward air temperature.**

Newton's Law of Cooling (1701):

$$Q_{conv} = h_{conv} \cdot (T_{surface} - T_{air})$$

The convective coefficient comes from empirical wind-over-flat-plate fits (Test et al. 1981; ASHRAE):

$$h_{conv} = 5.7 + 3.8 \cdot v \quad \text{(W/m²·K, v in m/s)}$$

At zero wind, h_conv ≈ 5.7 (natural convection from buoyant stirring of heated/cooled air). At 5 m/s (~11 mph), h_conv ≈ 24.7 — over 4× the still-air value. At 10 m/s, ~43.7.

**Wind dominates the cooling equation above a couple mph.** This is why:

- A hot south face becomes climbable earlier with a breeze.
- Cold windy days feel worse on rock (rock can't hold warmth).
- Calm, clear summer mornings produce the most dramatic temperature differentials.

> **Analogy:** Wind is blowing on hot soup. Faster wind, faster cooling toward air temperature.

---

## 6. Thermal Mass and Lag

The equation in section 2 gives T_eq, the **equilibrium** temperature — what the rock would settle at if conditions held constant. But conditions don't hold constant. Sun moves, clouds drift, wind shifts. The rock is always chasing a moving target.

Rock has thermal mass. Heating the top few cm of granite takes energy and time. The rock's surface acts as a flywheel resisting rapid change.

We model the climbable surface layer (top few cm — what your fingers touch) as a single lumped object with a thermal time constant τ:

$$T_{rock}(t + \Delta t) = T_{rock}(t) + (T_{eq}(t) - T_{rock}(t)) \cdot (1 - e^{-\Delta t / \tau})$$

After time τ the rock is 63% of the way to the new equilibrium; after 3τ, 95%; after 5τ, essentially there.

| Rock Type | τ | Why |
|-----------|---|-----|
| Sandstone | ~50 min | Porous, lower thermal mass per cm |
| Limestone | ~75 min | Medium density and porosity |
| Granite | ~105 min | Dense crystalline, high thermal mass |
| Basalt / Gabbro | ~105 min | Dense |

Values from heat-equation solutions for the top 3–5 cm with known thermal diffusivities (Robertson 1988; Cermák & Rybach 1982).

Consequences:

- **Peak rock temp lags peak air temp by 1–2 hours.** Hottest air at 3 PM; hottest granite at 4–5 PM.
- **Cold predawn rock warms slowly.** A 4 AM granite at 45°F doesn't hit 65°F until late morning.
- **Sandstone responds faster than granite to weather changes.** Storm cools sandstone in ~1 hour; granite stays warm 3+ hours.

> **Analogy:** Rock is a thermal flywheel. Air changes in minutes; rock takes hours to catch up.

### 6.1 Spin-Up

To know T_rock now, we need T_rock an hour ago. To know that, we need it two hours ago. And so on.

After ~5τ (8–9 hours for granite), any initial guess is fully forgotten. We pull the past 12 hours of weather (`past_hours=12` from Open-Meteo), initialize T_rock at the 12-hours-ago equilibrium, and integrate forward through past, present, and forecast. Without this, the first ~6 hours of forecast are biased by the initial guess.

---

## 7. Rock Types

Rock type matters in four distinct ways:

| Property | Effect | Driver |
|----------|--------|--------|
| Absorptivity (α) | Solar heat absorbed | Color and texture |
| Thermal mass (τ) | Response speed | Density and porosity |
| Friction-temp curve | What temp is "good" | Texture, edge geometry |
| Wet sensitivity | Whether dampness ruins friction | Mineral chemistry, porosity |

### 7.1 Granite

Coarse-grained igneous rock (quartz, feldspar, mica). Dense (~2,700 kg/m³), durable, visible crystals.

- **Color:** Light to medium gray. α ≈ 0.70.
- **Thermal mass:** High. τ ≈ 105 min.
- **Friction:** Highly temp-sensitive. The crystalline texture grips rubber via tiny edges; warm granite feels slick. Optimal surface temp: 35–55°F.
- **Wet sensitivity:** Low (though wet granite is slippery).
- **Examples:** Yosemite, Squamish, Joshua Tree, RMNP.

Granite's friction reputation is conditional — that friction only shows in cool temperatures. A Yosemite slab at 90°F in summer is a different animal from the same slab at 50°F in November.

### 7.2 Sandstone

Quartz grains cemented together. Lighter (~2,200 kg/m³), 5–25% porosity.

- **Color:** Highly variable — pale tan, red, varnished near-black. α 0.70–0.80, varnished up to 0.90.
- **Thermal mass:** Low-medium. τ ≈ 50 min.
- **Friction:** Abrasive — direct contact with thousands of quartz edges. Excellent in the cold; benefits most of any rock from low temps. Optimal: 30–50°F.
- **Wet sensitivity:** **HIGH.** Critical safety issue. Porous, absorbs water; wet cementing matrix weakens dramatically. Climbing wet sandstone breaks holds and damages the rock. Standard rule: **24–48 hours after rain on desert sandstone.**
- **Examples:** Indian Creek, Red Rock, Moe's, Joe's Valley.

Sandstone wet-sensitivity is so important it deserves its own subsystem. In the larger architecture, this is handled by the separate `rock_drying` package, not the temperature model.

### 7.3 Limestone

Calcium carbonate, often from marine sediments. Varies enormously in texture.

- **Color:** Light gray to nearly white. Lowest α of common climbing rocks: 0.55–0.65.
- **Thermal mass:** Medium. τ ≈ 75 min.
- **Friction:** Less temp-sensitive than granite or sandstone. Smooth, rounded surfaces grip more by hold shape than skin-to-rock friction; 70°F limestone still sends well. Optimal: 40–60°F, but the falloff at higher temps is gentler.
- **Wet sensitivity:** Medium. Slick when wet but not structurally damaged.
- **Examples:** Rifle, American Fork, Wild Iris, Kalymnos.

### 7.4 Basalt and Gabbro

Dark, dense igneous rocks. Basalt is fine-grained (surface lava); gabbro is coarse-grained (slow-cooled underground). Climbing behavior is similar.

- **Color:** Dark gray to black. Highest α: 0.85–0.95.
- **Thermal mass:** High. τ ≈ 105 min.
- **Friction:** Variable — some basalt is rough and sticky like granite; some is glassy and slick. Optimal: 35–55°F.
- **Wet sensitivity:** Low.
- **Examples:** Columbia River Gorge, parts of Tahoe.

The combination of high absorptivity + high thermal mass makes basalt the worst rock for summer climbing. It heats aggressively and won't give up the heat. A south-facing basalt cliff in July is essentially uncategorizable until well into the evening.

### 7.5 Quartzite

Metamorphosed sandstone — pressure-cooked over geological time into dense, hard rock of fused quartz grains.

- **Color:** Light gray to white. α 0.65–0.75.
- **Thermal mass:** High. τ ≈ 105 min.
- **Friction:** Similar to granite, slightly more heat-tolerant due to very fine crystalline texture. Optimal: 35–58°F.
- **Wet sensitivity:** Low (the metamorphism welded the grains).
- **Examples:** Pilot Mountain (NC), Devil's Lake (WI).

### 7.6 Rock Type Groups

In the database, rock types group into families (`rock_type_groups`). The calculator keys off the **group**, not the specific rock type:

- "Yosemite Granite" and "Sierra Granite" → granite group → granite physics
- "Wingate Sandstone" and "Navajo Sandstone" → sandstone group → sandstone physics

This reflects how climbers actually think about conditions — by family, not subtype.

---

## 8. Friction Quality Thresholds

The temperature-to-condition mapping is part physics, part empirical. Physics: rubber and skin friction degrade with rising temperature (rubber softens and glazes; skin sweats and loses grip). Empirical: generations of climbers logging conditions on their best and worst sends.

### 8.1 Tiers

| Tier | Meaning |
|------|---------|
| **Prime** | Elite-level friction. Send temps. |
| **Good** | Most climbers won't notice degradation. |
| **Marginal** | Projecting suffers; warmups fine. |
| **Poor** | Fighting the rock; hard moves feel impossible. |
| **Very Poor** | Unclimbable for performance. |
| **Too Cold** | Numb skin, no feedback, skin splits. |

### 8.2 Thresholds by Rock Type (Surface °F)

| Rock | Prime | Good | Marginal | Poor | Very Poor | Too Cold |
|------|-------|------|----------|------|-----------|----------|
| Granite | 35–55 | 55–65 | 65–72 | 72–85 | >85 | <30 |
| Sandstone | 30–50 | 50–60 | 60–68 | 68–80 | >80 | <25 |
| Basalt/Gabbro | 35–55 | 55–63 | 63–70 | 70–85 | >85 | <30 |
| Limestone | 40–60 | 60–68 | 68–75 | 75–85 | >85 | <30 |
| Quartzite | 35–58 | 58–67 | 67–74 | 74–85 | >85 | <30 |

Synthesized from climbing community knowledge — bouldering forums, European friction literature (Güllich-era), Mountain Project and 8a.nu condition discussions. Calibrated to surface temperature, not air.

### 8.3 Why Thresholds Differ

**Sandstone wants colder than granite.** Two reasons: sandstone's friction comes from grinding many small quartz edges into skin, and cooler skin sweats less and contacts more crisply. Also, sandstone's low thermal mass means it warms quickly from body heat at the hold — you get "one shot" before local friction degrades. Starting colder buys you more time.

**Limestone tolerates heat better.** Climbing leans more on hold shape than skin friction. A jug is a jug at 50°F or 70°F.

**"Too cold" is a real category** because below ~30°F skin loses tactile feedback (can't tell positive from greasy holds) and splits more easily (cold skin is brittle). Sandstone's threshold is 25°F because its gentler texture is easier on skin.

---

## 9. Condensation

Even at prime temperature, condensation can ruin friction. This happens when rock surface temperature drops below the air's dewpoint and water vapor condenses onto the rock.

### 9.1 Dewpoint Mechanics

**Dewpoint** is the temperature at which air, cooled at constant moisture, saturates. Air at 70°F with 55°F dewpoint will start condensing if cooled to 55°F. Any *surface* below 55°F sitting in that air will collect condensation.

Most common scenario: **early morning at the crag.** Overnight radiative cooling drops rock 10°F below air. Sunrise warms air faster than rock (air heats by convection from warm surroundings; rock has to integrate through thermal mass). Result: air at 60°F with 52°F dewpoint, rock at 50°F. Rock is below dewpoint and weeping moisture.

This is why 7 AM sessions feel inexplicably slippery in "perfect" temps. The friction degradation is real — microscopic water on the rock — and waiting until 9 AM transforms the session.

Other scenarios: cold fronts with humid air over cold rock; fog/marine layer keeping dewpoint at air temp; wind cooling rock below air while humidity stays high.

### 9.2 Severity Classification

We classify condensation by T_rock − T_dewpoint:

| Difference | Severity | What it feels like |
|------------|----------|--------------------|
| > +2°F | None | Dry rock |
| 0 to +2°F | Light | Damp microclimate, humidity collecting |
| < 0°F | Heavy | Visibly wet, unclimbable |

The 2°F buffer reflects that air close to the rock can be more humid than the bulk forecast (the rock releases or absorbs moisture). For "clears at" predictions, we look forward for the first hour where T_rock ≥ T_dewpoint + 2°F — the extra margin prevents the model from flapping between wet/dry near the boundary.

### 9.3 Friction Combination

We combine temperature condition and condensation severity into a single friction rating:

- **Heavy condensation overrides everything.** Wet rock is poor regardless of temp.
- **Light condensation drops by one tier.** Prime + light → reduced.
- **Dry rock + temp tier** → friction maps directly.

This is what the UI badge surfaces.

---

## 10. Worked Example

**Scenario:** South-facing granite wall at 2 PM in late August, Sierra Nevada, elevation 8,500 ft.

**Inputs:**
- T_air = 75°F (24°C), cloud cover = 10%, wind = 3 mph (1.3 m/s)
- Direct radiation (horizontal) = 700 W/m², diffuse = 120 W/m²
- Sun elevation = 55°, azimuth = 215°
- Face aspect = 180° (due south), dip = 90° (vertical)
- Tree coverage = 10%
- Granite: α = 0.70, τ = 105 min
- Dewpoint = 45°F

**Step 1: DNI conversion**
I_DNI = 700 / sin(55°) = 700 / 0.819 = 855 W/m²

**Step 2: Geometric factor**
cos(θ) = cos(55°) × cos(215° − 180°) = 0.574 × 0.819 = 0.470

**Step 3: Effective irradiance on face**
I_face = 855 × 0.470 × (1 − 0.10×0.75) + 120 × 0.5 × (1 − 0.10×0.5)
     = 372 + 57 = 429 W/m²

(F_sky = (1 + cos(90°))/2 = 0.5 for the vertical wall.)

**Step 4: Elevation correction**
At 8,500 ft (2.59 km), irradiance multiplier = 1 + 0.05×2.59 = 1.13. I_face becomes 484 W/m².

**Step 5: Convective coefficient**
h_conv = 5.7 + 3.8 × 1.3 = 10.6 W/m²·K

**Step 6: Sky temperature**
T_sky = T_air − 20°F × (1 − 0.10) = 75 − 18 = 57°F. In Celsius: ΔT_sky = 10°C.

**Step 7: Energy balance**
- Solar heating: α × I_face = 0.70 × 484 = 339 W/m²
- Radiative cooling: h_rad × ΔT_sky = 5.5 × 10 = 55 W/m²
- Net numerator: 339 − 55 = 284 W/m²
- Denominator: h_conv + h_rad = 10.6 + 5.5 = 16.1 W/m²·K
- ΔT = 284 / 16.1 = 17.6°C = 31.7°F

**Step 8: Equilibrium temperature**
T_eq = 75°F + 31.7°F = 106.7°F

If conditions have held for hours: T_rock ≈ 107°F.

**Step 9: Condition**
For granite, >85°F surface = "very_poor." Skip.

**Step 10: When does the send window open?**

Running the same calculation forward each hour as sun drops. Around 6 PM (sun elev ≈ 15°), the geometric factor shrinks, α × I_face becomes ~50 W/m². By 8 PM with the sun gone, accounting for sky radiation:

T_eq ≈ 69°F

But rock has thermal mass. From 107°F at 4 PM toward 69°F target, with τ = 105 min:
- 1 hr later (5 PM): T_rock = 107 + (69−107) × (1−e^(−60/105)) = 90.5°F
- 2 hrs (6 PM): ~80°F
- 3 hrs (7 PM): ~74°F
- 4 hrs (8 PM): ~71°F

Send window opens around 8–9 PM at ~70°F (still "marginal" for granite). Good territory by 10–11 PM. Prime by 1–2 AM as overnight sky radiation kicks in.

This matches what Sierra climbers actually do for summer projects: night sessions, best burns between midnight and dawn.

---

## 11. Confidence and Limitations

The `confidence_score` (0–100) reflects how much to trust the prediction. Things that reduce confidence:

- **Unknown aspect** (−25): location's aspect not recorded; default is "average" which can be wildly off for individual walls.
- **Unknown dip** (−15): defaulting to vertical is wrong for slabs and overhangs.
- **Forecast distance** (−5/day): forecasts degrade with time.
- **Variable cloud cover** (−10): broken clouds make irradiance forecasts unreliable.
- **Variable wind** (−10): gusty conditions throw off convection estimates.
- **Unknown rock type** (−5): defaulting to granite is reasonable but imprecise.
- **Missing spin-up** (−5): incomplete past data means the model is still warming up.

### Known Limitations

- **Terrain shading.** A wall may go into shadow from a neighboring ridge well before astronomical sunset. The model doesn't know about local terrain.
- **Microclimates.** Valley inversions, water-induced humidity boosts, canyon-wall heat reflection.
- **Color variation within a wall.** Desert varnish streaks have different absorptivity than surrounding rock.
- **Crag-scale wind variability.** Forecast wind is from airports or coarse grids; actual wind at a boulder in a notch can be very different.
- **Long-term thermal history.** Spin-up handles 12 hours; deeper seasonal effects (e.g., first session after months of shadow) aren't captured.

### Out of Scope

- **Safety judgments.** Rock at 90°F isn't dangerous, just bad climbing. The model predicts friction, not risk.
- **Rock-drying after rain.** Handled by the `rock_drying` package, not this model.
- **Whether you'll send.** That depends on you.

---

## 12. Sources

### Primary Physics

- Bergman, Lavine, Incropera, DeWitt. *Fundamentals of Heat and Mass Transfer*, 7th ed. Wiley, 2011. — conduction, convection, radiation basics; linearization of T⁴.
- Newton, I. *Philosophical Transactions of the Royal Society*, 1701. — Law of Cooling.
- Lambert, J.H. *Photometria*, 1760. — cosine law for incident light.
- Test, F.L., Lessmann, R.C., Johary, A. "Heat transfer during wind flow over rectangular bodies in the natural environment," *J. Heat Transfer* 103, 1981. — convective coefficient correlation.
- Berdahl, P., Fromberg, R. "The thermal radiance of clear skies," *Solar Energy* 29(4), 1982. — clear-sky effective temperature.
- Swinbank, W.C. "Long-wave radiation from clear skies," *Quarterly Journal of the Royal Meteorological Society* 89, 1963.

### Rock Properties

- Robertson, E.C. *Thermal Properties of Rocks.* USGS Open-File Report 88-441, 1988.
- Cermák, V., Rybach, L. "Thermal properties: thermal conductivity and specific heat of minerals and rocks" in *Landolt-Börnstein*, Springer, 1982.
- ASHRAE Handbook of Fundamentals — surface property tables.
- Clark, S.P. (ed.) *Handbook of Physical Constants.* Geological Society of America Memoir 97, 1966.
- Dorn, R.I. *Rock Coatings.* Elsevier, 1998. — desert varnish absorptivity changes.

### Solar Geometry

- Reda, I., Andreas, A. "Solar position algorithm for solar radiation applications," NREL TP-560-34302, 2008.
- Erbs, D.G., Klein, S.A., Duffie, J.A. "Estimation of the diffuse radiation fraction for hourly, daily and monthly-average global radiation," *Solar Energy* 28(4), 1982.

### Forest Microclimate

- Federer, C.A. "Solar radiation absorption by leafless hardwood forests," *Agricultural Meteorology* 9, 1971.
- Hardy, J.P. et al. "Solar radiation transmission through conifer canopies," *Agricultural and Forest Meteorology* 126, 2004.

### Data

- Open-Meteo Forecast API: https://open-meteo.com/en/docs
- Open-Meteo Historical Weather API: https://open-meteo.com/en/docs/historical-weather-api

### Climbing Community

- Mountain Project forums — climbing condition discussions.
- 8a.nu user comments on send conditions.
- Area-specific bouldering guides (Hueco, RMNP, Joe's Valley, etc.) documenting local temperature preferences.
- Climbing rubber friction-vs-temperature data is proprietary to manufacturers (La Sportiva, Five Ten, Scarpa); precise curves aren't published, but general trends are well-known.