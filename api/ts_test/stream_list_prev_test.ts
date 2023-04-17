import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream list prev success', async () => {
	const tc = new TestClient();

	await tc.client.createTestType({text: "foo"});
	await tc.client.createTestType({text: "bar"});

	const list = await tc.client.listTestType();
	assert.equal(list!.length, 2);

	// This is test-only
	// Don't mutate objects and pass them back in GetOpts.prev
	list[0]!.num = 5;

	const stream = await tc.client.streamListTestType({prev: list});

	try {
		const s1 = await stream.read();
		assert(s1);
		assert.equal(s1.length, 2);
		assert.equal(s1[0]!.num, 5);
	} finally {
		await stream.close();
	}
});
