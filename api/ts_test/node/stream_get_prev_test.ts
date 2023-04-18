import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream get prev success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: 'foo'});

	// This is test-only
	// Don't mutate objects and pass them back in GetOpts.prev
	create.num = 5;

	const stream = await tc.client.streamGetTestType(create.id, {prev: create});

	try {
		const s1 = await stream.read();
		assert.equal(s1!.text, 'foo');
		assert.equal(s1!.num, 5);
	} finally {
		await stream.close();
	}
});
