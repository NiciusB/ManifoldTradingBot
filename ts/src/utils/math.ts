/**
 * Maps a number from one range to another
 * @param value The value to map
 * @param inMin The minimum value of the input range
 * @param inMax The maximum value of the input range
 * @param outMin The minimum value of the output range
 * @param outMax The maximum value of the output range
 * @returns The mapped value
 */
export function mapNumber(
  value: number,
  inMin: number,
  inMax: number,
  outMin: number,
  outMax: number,
): number {
  return ((value - inMin) * (outMax - outMin)) / (inMax - inMin) + outMin;
}

export function probToOdds(prob: number): number {
  return prob / (1 - prob);
}

export function oddsToProb(odds: number): number {
  return odds / (1 + odds);
}
