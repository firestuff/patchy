import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream get success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: "foo"});

	// This is test-only
	// Don't mutate objects and pass them back in GetOpts.prev
	create.num = 5;

	const stream = await tc.client.streamGetTestType(create.id, {prev: create});

	const ev = await stream.read();
	assert.equal(ev!.obj.text, "foo");
	assert.equal(ev!.obj.num, 5);

	await stream.abort();
	assert.rejects(stream.read());
});
