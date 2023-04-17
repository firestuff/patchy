import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('update success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: 'foo', num: 5});
	const get1 = await tc.client.getTestType(create.id);
	await tc.client.updateTestType(create.id, {text: 'bar'});

	assert.rejects(tc.client.updateTestType(create.id, {text: 'zig'}, {prev: get1}));

	const get2 = await tc.client.getTestType(create.id);
	assert.equal(get2.text, 'bar');
});
