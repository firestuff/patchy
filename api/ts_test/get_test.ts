import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

const tc = new TestClient();

const create = await tc.client.createTestType({text: "foo"});

const get1 = await tc.client.getTestType(create.id);
assert.equal(get1.text, "foo");

// TODO: Mutate get1 to be sneaky

const get2 = await tc.client.getTestType(create.id, {prev: get1});
assert.equal(get2.text, "foo");
