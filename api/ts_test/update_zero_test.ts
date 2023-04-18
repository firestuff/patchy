import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('update zero success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: 'foo', num: 5});

	await tc.client.updateTestType(create.id, {num: 0});

	const get = await tc.client.getTestType(create.id);
	assert.equal(get.text, 'foo');
	assert.equal(get.num, 0);
});
