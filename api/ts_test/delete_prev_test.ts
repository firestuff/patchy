import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('delete prev failure', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: 'foo'});
	const get = await tc.client.getTestType(create.id);
	await tc.client.updateTestType(create.id, {text: 'bar'});

	assert.rejects(tc.client.deleteTestType(create.id, {prev: get}));

	await tc.client.getTestType(create.id);
});
