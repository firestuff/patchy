import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('get prev success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: 'foo'});

	const get1 = await tc.client.getTestType(create.id);
	assert.equal(get1.text, 'foo');

	// This is test-only
	// Don't mutate objects and pass them back in GetOpts.prev
	get1.num = 5;

	const get2 = await tc.client.getTestType(create.id, {prev: get1});
	assert.equal(get2.text, 'foo');
	assert.equal(get2.num, 5);
});
