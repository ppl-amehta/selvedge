import test from 'node:test';
import assert from 'node:assert/strict';
import {
  normInwardKind,
  normalizeNonApparelGenre,
  defaultUomForInward,
  getProductOptionsForInward,
  getAllowedSizesForInward,
  getPackagingSizes,
  getPackagingVariants,
  getTrimSizes,
  getZipTeethSizes,
  reconcileBundleSizeSpec
} from './inwardRules.js';

const mockMasters = {
  inward_types: [
    { display_name: 'Fabric' },
    { display_name: 'Zips' },
    { display_name: 'packaging' },
    { display_name: 'Trims' },
    { display_name: 'Label' },
    { display_name: 'Apparel' },
    { display_name: 'Drinkwear' },
    { display_name: 'Bags' },
    { display_name: 'Headgear' },
    { display_name: 'Electronic' },
    { display_name: 'Stationary' },
    { display_name: 'Other Inwards' }
  ],
  fabrics: [
    { display_name: '2441 Galactic Polo', meta: {} },
    { display_name: '30s Sinker Crew Bio Wash', meta: {} }
  ],
  products: [
    { display_name: 'Mens Classic Polo Shirt', meta: { category: 'Apparel', subcategory: 'Polo' } },
    { display_name: 'Heavyweight Fleece Hoodie', meta: { category: 'Apparel', subcategory: 'Hoodie' } },
    { display_name: 'Round Neck Crew Tee', meta: { category: 'Apparel', subcategory: 'Tee' } },
    { display_name: 'Stainless Vacuum Flask 500ml', meta: { category: 'Non Apparel', subcategory: 'Drinkwear' } },
    { display_name: 'Canvas Tote Bag', meta: { category: 'Non Apparel', subcategory: 'Bags' } },
    { display_name: '6-Panel Structured Cap', meta: { category: 'Non Apparel', subcategory: 'Headgear' } },
    { display_name: 'MagSafe Power Bank', meta: { category: 'Non Apparel', subcategory: 'Electronics' } }
  ],
  gender_sizes: {
    Men: ['S', 'M', 'L', 'XL'],
    Infant: ['0-3', '3-6', '6-9'],
    Kids: ['K16', 'K18', 'K20']
  },
  non_apparel_sizes: {
    Drinkwear: ['100ml', '500ml', '1Litre'],
    Headgear: ['Std', 'Kids', 'Large', 'Small'],
    Bags: ['Std'],
    Stationary: ['A4', 'A5'],
    Electronics: ['Std'],
    Others: ['Std']
  },
  zip_sizes: [
    { display_name: '10.0' },
    { display_name: '5.0' },
    { display_name: '8.0' },
    { display_name: 'na' }
  ],
  zip_types: [{ display_name: 'Metal' }, { display_name: 'Coil' }],
  zip_closures: [{ display_name: 'open ended' }, { display_name: 'closed ended' }],
  trim_button_sizes: [{ display_name: '14L' }, { display_name: '18L' }],
  trim_elastic_sizes: [{ display_name: 'Elastic Size' }, { display_name: '1"' }],
  trim_thread_sizes: [{ display_name: 'Tex 30' }, { display_name: 'Tex 40' }],
  packaging_types: [{ display_name: 'size' }, { display_name: 'Poly Bag Self seal' }, { display_name: 'Hangtag' }],
  packaging_polybag_sizes: [{ display_name: '12x16"' }],
  packaging_polybag_brands: [{ display_name: 'Ettu' }, { display_name: 'Blank' }],
  packaging: [
    { display_name: 'FPCC Carton' },
    { display_name: 'Poly Bag Self seal' },
    { display_name: 'ETTU Hangtag' }
  ],
  label_brands: [{ display_name: 'ETTU' }],
  label_additionals: [{ display_name: 'washcare Label' }],
  label_genres: [{ display_name: 'Woven' }, { display_name: 'Printed' }]
};

test('normInwardKind distinguishes unset, non_apparel, packaging, and other without defaulting empty to fabric', () => {
  assert.equal(normInwardKind(''), 'unset');
  assert.equal(normInwardKind('Fabric'), 'fabric');
  assert.equal(normInwardKind('packaging'), 'packaging');
  assert.equal(normInwardKind('Drinkwear'), 'non_apparel');
  assert.equal(normInwardKind('Electronic'), 'non_apparel');
  assert.equal(normInwardKind('Stationary'), 'non_apparel');
  assert.equal(normInwardKind('Other Inwards'), 'other');
});

test('getProductOptionsForInward prevents cross-category product inwarding', () => {
  const drinkwear = getProductOptionsForInward(mockMasters, 'Drinkwear');
  assert.deepEqual(drinkwear.map(x => x.value), ['Stainless Vacuum Flask 500ml']);

  const electronics = getProductOptionsForInward(mockMasters, 'Electronic');
  assert.deepEqual(electronics.map(x => x.value), ['MagSafe Power Bank']);

  const bags = getProductOptionsForInward(mockMasters, 'Bags');
  assert.deepEqual(bags.map(x => x.value), ['Canvas Tote Bag']);

  const polos = getProductOptionsForInward(mockMasters, 'Apparel', { genre: 'Polo' });
  assert.deepEqual(polos.map(x => x.value), ['Mens Classic Polo Shirt']);

  const hoodies = getProductOptionsForInward(mockMasters, 'Apparel', { genre: 'Hoodies and Pullovers' });
  assert.deepEqual(hoodies.map(x => x.value), ['Heavyweight Fleece Hoodie']);

  const stationary = getProductOptionsForInward(mockMasters, 'Stationary');
  assert.equal(stationary.length, 0);

  const zips = getProductOptionsForInward(mockMasters, 'Zips', { zipType: 'Metal', teethSize: '5.0', closure: 'open ended' });
  assert.ok(zips.some(z => z.value === 'Metal #5.0 open ended'));
  assert.ok(!zips.some(z => z.value.includes('Polo')));
});

test('getAllowedSizesForInward isolates sizes per inward type and gender', () => {
  assert.deepEqual(getAllowedSizesForInward(mockMasters, 'Drinkwear'), ['100ml', '500ml', '1Litre']);
  assert.deepEqual(getAllowedSizesForInward(mockMasters, 'Electronic'), ['Std']);
  assert.deepEqual(getAllowedSizesForInward(mockMasters, 'Stationary'), ['A4', 'A5']);
  assert.deepEqual(getAllowedSizesForInward(mockMasters, 'Apparel', { gender: 'Kids' }), ['K16', 'K18', 'K20']);
  assert.deepEqual(getAllowedSizesForInward(mockMasters, 'Apparel', { gender: 'Infant' }), ['0-3', '3-6', '6-9']);

  const allApparel = getAllowedSizesForInward(mockMasters, 'Apparel', { gender: 'All Apparel' });
  assert.ok(!allApparel.includes('500ml'));
  assert.ok(!allApparel.includes('A4'));

  assert.deepEqual(getZipTeethSizes(mockMasters), ['5.0', '8.0', '10.0', 'na']);
  assert.deepEqual(getTrimSizes(mockMasters, 'Elastic'), ['1"']);
  assert.deepEqual(getTrimSizes(mockMasters, 'Thread'), ['Tex 30', 'Tex 40']);
  assert.deepEqual(getPackagingSizes(mockMasters, 'Poly Bag Self seal'), ['12x16"']);
  assert.deepEqual(getPackagingVariants(mockMasters, 'Poly Bag Self seal'), ['Ettu', 'Blank']);
  assert.ok(getPackagingSizes(mockMasters, 'Hangtag').includes('Std'));
});

test('reconcileBundleSizeSpec scrubs incompatible sizes across inward type changes', () => {
  assert.equal(reconcileBundleSizeSpec('500ml', ['14L', '18L'], '14L'), '14L');
  assert.equal(reconcileBundleSizeSpec('18L', ['14L', '18L'], '14L'), '18L');
  assert.equal(defaultUomForInward('Fabric'), 'Kgs');
  assert.equal(defaultUomForInward('Trims', 'Thread'), '1000 mt tube');
  assert.equal(defaultUomForInward('Drinkwear'), 'Pcs');
  assert.equal(normalizeNonApparelGenre('Electronic'), 'Electronics');
});
