import { describe, it, expect } from 'vitest';
import {
  gradeFamily,
  gradeToOrderNum,
  orderToGrade,
  getGradeScalesForTypes,
  gradeRangeToApiParams,
  FAMILY_V,
  FAMILY_YDS,
  FAMILY_WI,
  FAMILY_AI,
  FAMILY_MIXED,
  FAMILY_AID,
  V_SCALE_GRADES,
  YDS_GRADES,
  WI_GRADES,
  AI_GRADES,
  MIXED_GRADES,
} from '../../utils/grades';

describe('gradeFamily', () => {
  it('classifies V-scale grades', () => {
    expect(gradeFamily('V0')).toBe(FAMILY_V);
    expect(gradeFamily('V10')).toBe(FAMILY_V);
    expect(gradeFamily('v4')).toBe(FAMILY_V);
  });

  it('classifies YDS grades', () => {
    expect(gradeFamily('5.10a')).toBe(FAMILY_YDS);
    expect(gradeFamily('5.9')).toBe(FAMILY_YDS);
    expect(gradeFamily('5.12d')).toBe(FAMILY_YDS);
  });

  it('classifies WI grades', () => {
    expect(gradeFamily('WI3')).toBe(FAMILY_WI);
    expect(gradeFamily('wi5')).toBe(FAMILY_WI);
  });

  it('classifies AI (Alpine Ice) grades', () => {
    expect(gradeFamily('AI3')).toBe(FAMILY_AI);
    expect(gradeFamily('ai5')).toBe(FAMILY_AI);
    expect(gradeFamily('AI1')).toBe(FAMILY_AI);
  });

  it('classifies Mixed grades', () => {
    expect(gradeFamily('M5')).toBe(FAMILY_MIXED);
    expect(gradeFamily('m10')).toBe(FAMILY_MIXED);
  });

  it('classifies Aid grades', () => {
    expect(gradeFamily('A0')).toBe(FAMILY_AID);
    expect(gradeFamily('C2')).toBe(FAMILY_AID);
  });

  it('returns undefined for unrecognized grades', () => {
    expect(gradeFamily('')).toBeUndefined();
    expect(gradeFamily('unknown')).toBeUndefined();
  });
});

describe('gradeToOrderNum', () => {
  it('converts V-scale grades', () => {
    expect(gradeToOrderNum('V0')).toBe(0);
    expect(gradeToOrderNum('V4')).toBe(4);
    expect(gradeToOrderNum('V17')).toBe(17);
  });

  it('converts YDS grades', () => {
    expect(gradeToOrderNum('5.4')).toBe(100);
    expect(gradeToOrderNum('5.9')).toBe(105);
    expect(gradeToOrderNum('5.10a')).toBe(106);
    expect(gradeToOrderNum('5.12d')).toBe(117);
  });

  it('converts WI grades', () => {
    expect(gradeToOrderNum('WI1')).toBe(200);
    expect(gradeToOrderNum('WI7')).toBe(206);
  });

  it('converts AI grades', () => {
    expect(gradeToOrderNum('AI1')).toBe(400);
    expect(gradeToOrderNum('AI6')).toBe(405);
  });

  it('converts Mixed grades', () => {
    expect(gradeToOrderNum('M1')).toBe(300);
    expect(gradeToOrderNum('M13')).toBe(312);
  });

  it('handles case insensitivity', () => {
    expect(gradeToOrderNum('v4')).toBe(4);
    expect(gradeToOrderNum('wi3')).toBe(202);
    expect(gradeToOrderNum('5.10A')).toBe(106);
  });

  it('handles grades with +/- modifiers', () => {
    expect(gradeToOrderNum('V4+')).toBe(4);
    expect(gradeToOrderNum('V4-')).toBe(4);
  });

  it('handles slash grades', () => {
    expect(gradeToOrderNum('5.10a/b')).toBe(106);
  });

  it('handles bare YDS without letter', () => {
    expect(gradeToOrderNum('5.10')).toBe(106); // maps to 5.10a
    expect(gradeToOrderNum('5.11')).toBe(110); // maps to 5.11a
  });

  it('returns -1 for unrecognized grades', () => {
    expect(gradeToOrderNum('')).toBe(-1);
    expect(gradeToOrderNum('unknown')).toBe(-1);
  });
});

describe('orderToGrade', () => {
  it('converts V-scale orders', () => {
    expect(orderToGrade(0)).toBe('V0');
    expect(orderToGrade(4)).toBe('V4');
    expect(orderToGrade(17)).toBe('V17');
  });

  it('converts YDS orders', () => {
    expect(orderToGrade(100)).toBe('5.4');
    expect(orderToGrade(106)).toBe('5.10a');
  });

  it('converts WI orders', () => {
    expect(orderToGrade(200)).toBe('WI1');
  });

  it('converts AI orders', () => {
    expect(orderToGrade(400)).toBe('AI1');
    expect(orderToGrade(405)).toBe('AI6');
  });

  it('converts Mixed orders', () => {
    expect(orderToGrade(300)).toBe('M1');
  });

  it('returns undefined for invalid orders', () => {
    expect(orderToGrade(-1)).toBeUndefined();
    expect(orderToGrade(999)).toBeUndefined();
  });
});

describe('grade ordering consistency', () => {
  it('V-scale grades are ordered correctly', () => {
    for (let i = 1; i < V_SCALE_GRADES.length; i++) {
      expect(gradeToOrderNum(V_SCALE_GRADES[i])).toBeGreaterThan(
        gradeToOrderNum(V_SCALE_GRADES[i - 1]),
      );
    }
  });

  it('YDS grades are ordered correctly', () => {
    for (let i = 1; i < YDS_GRADES.length; i++) {
      expect(gradeToOrderNum(YDS_GRADES[i])).toBeGreaterThan(
        gradeToOrderNum(YDS_GRADES[i - 1]),
      );
    }
  });

  it('WI grades are ordered correctly', () => {
    for (let i = 1; i < WI_GRADES.length; i++) {
      expect(gradeToOrderNum(WI_GRADES[i])).toBeGreaterThan(
        gradeToOrderNum(WI_GRADES[i - 1]),
      );
    }
  });

  it('Mixed grades are ordered correctly', () => {
    for (let i = 1; i < MIXED_GRADES.length; i++) {
      expect(gradeToOrderNum(MIXED_GRADES[i])).toBeGreaterThan(
        gradeToOrderNum(MIXED_GRADES[i - 1]),
      );
    }
  });

  it('AI grades are ordered correctly', () => {
    for (let i = 1; i < AI_GRADES.length; i++) {
      expect(gradeToOrderNum(AI_GRADES[i])).toBeGreaterThan(
        gradeToOrderNum(AI_GRADES[i - 1]),
      );
    }
  });

  it('round-trips through orderToGrade for every grade', () => {
    for (const grade of [...V_SCALE_GRADES, ...YDS_GRADES, ...WI_GRADES, ...AI_GRADES, ...MIXED_GRADES]) {
      const order = gradeToOrderNum(grade);
      expect(order).toBeGreaterThanOrEqual(0);
      expect(orderToGrade(order)).toBe(grade);
    }
  });
});

describe('getGradeScalesForTypes', () => {
  it('returns boulder scale when Boulder is selected', () => {
    const scales = getGradeScalesForTypes(['Boulder']);
    expect(scales).toHaveLength(1);
    expect(scales[0].key).toBe('boulder');
    expect(scales[0].grades).toEqual(V_SCALE_GRADES);
  });

  it('returns rope scale when Sport is selected', () => {
    const scales = getGradeScalesForTypes(['Sport']);
    expect(scales).toHaveLength(1);
    expect(scales[0].key).toBe('rope');
    expect(scales[0].label).toBe('Sport');
  });

  it('combines Sport/Trad into single rope scale', () => {
    const scales = getGradeScalesForTypes(['Sport', 'Trad']);
    expect(scales).toHaveLength(1);
    expect(scales[0].label).toBe('Sport / Trad');
  });

  it('returns ice and mixed scales when Ice is selected', () => {
    const scales = getGradeScalesForTypes(['Ice']);
    expect(scales).toHaveLength(2);
    expect(scales[0].key).toBe('ice');
    expect(scales[0].grades.length).toBe(WI_GRADES.length + AI_GRADES.length);
    expect(scales[1].key).toBe('mixed');
    expect(scales[1].grades.length).toBe(MIXED_GRADES.length);
  });

  it('returns multiple scales when multiple types selected', () => {
    const scales = getGradeScalesForTypes(['Boulder', 'Sport', 'Ice']);
    expect(scales).toHaveLength(4);
    expect(scales.map((s) => s.key)).toEqual(['boulder', 'rope', 'ice', 'mixed']);
  });

  it('returns empty for no types', () => {
    expect(getGradeScalesForTypes([])).toHaveLength(0);
  });
});

describe('gradeRangeToApiParams', () => {
  it('returns empty when all at full range', () => {
    const result = gradeRangeToApiParams(
      { boulder: [0, V_SCALE_GRADES.length - 1] },
      ['Boulder'],
    );
    expect(result).toEqual({});
  });

  it('returns grade orders when range is narrowed', () => {
    const result = gradeRangeToApiParams(
      { boulder: [3, 5] },
      ['Boulder'],
    );
    // V3=3, V4=4, V5=5
    expect(result.gradeOrders).toBe('3,4,5');
  });

  it('returns YDS orders for rope scale', () => {
    const result = gradeRangeToApiParams(
      { rope: [5, 7] },
      ['Sport'],
    );
    // YDS offset=100, indices 5-7 → orders 105,106,107
    expect(result.gradeOrders).toBe('105,106,107');
  });

  it('includes full-range orders from unfiltered families', () => {
    const result = gradeRangeToApiParams(
      {
        boulder: [2, 5],
        // rope not set → full range → include ALL rope orders
      },
      ['Boulder', 'Sport'],
    );
    // Boulder narrowed: V2=2, V3=3, V4=4, V5=5
    // Rope full range: all YDS orders (100..147)
    const orders = result.gradeOrders!.split(',').map(Number);
    // Check boulder part
    expect(orders).toContain(2);
    expect(orders).toContain(5);
    expect(orders).not.toContain(0); // V0 excluded
    expect(orders).not.toContain(1); // V1 excluded
    // Check rope part (full range included)
    expect(orders).toContain(100); // 5.4
    expect(orders).toContain(147); // 5.15d
  });

  it('includes orders from multiple narrowed families', () => {
    const result = gradeRangeToApiParams(
      {
        boulder: [0, 1],
        rope: [0, 1],
      },
      ['Boulder', 'Sport'],
    );
    // V0=0, V1=1, 5.4=100, 5.5=101
    expect(result.gradeOrders).toBe('0,1,100,101');
  });

  it('returns empty when no selections provided', () => {
    expect(gradeRangeToApiParams({}, ['Boulder'])).toEqual({});
  });
});
