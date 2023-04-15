import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream get success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: "foo"});

	const stream = await tc.client.streamGetTestType(create.id);

	const ev1 = await stream.read();
	assert.equal(ev1!.obj.text, "foo");

	await tc.client.updateTestType(create.id, {text: "bar"});

	const ev2 = await stream.read();
	assert.equal(ev2!.obj.text, "bar");

	await stream.abort();
	assert.rejects(stream.read());
});