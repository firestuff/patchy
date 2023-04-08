import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

const tc = new TestClient();

const create = await tc.client.createTestType({text: "foo"});
assert.equal(create.text!, "foo");

const get1 = await tc.client.getTestType(create.id);
assert.equal(get1.text!, "foo");

const update = await tc.client.updateTestType(create.id, {text: "bar"});
assert.equal(update.text!, "bar");

const get2 = await tc.client.getTestType(create.id);
assert.equal(get2.text!, "bar");
