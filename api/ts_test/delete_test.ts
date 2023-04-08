import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('delete success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: "foo"});
	assert.equal(create.text, "foo");

	const get = await tc.client.getTestType(create.id);
	assert.equal(get.text, "foo");

	await tc.client.deleteTestType(create.id);

	assert.rejects(tc.client.getTestType(create.id));
});
