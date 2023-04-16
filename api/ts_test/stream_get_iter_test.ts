import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream get success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: "foo"});

	const stream = await tc.client.streamGetTestType(create.id);

	await tc.client.updateTestType(create.id, {text: "bar"});

	const evs = [];

	for await (const ev of stream) {
		evs.push(ev);

		if (evs.length == 2) {
			stream.abort();
		}
	}

	assert.equal(evs.length, 2);
	assert.equal(evs[0]!.text, "foo");
	assert.equal(evs[1]!.text, "bar");
});
