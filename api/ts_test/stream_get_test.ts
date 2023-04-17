import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream get success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: "foo"});

	const stream = await tc.client.streamGetTestType(create.id);

	try {
		const s1 = await stream.read();
		assert.equal(s1!.text, "foo");

		await tc.client.updateTestType(create.id, {text: "bar"});

		const s2 = await stream.read();
		assert.equal(s2!.text, "bar");
	} finally {
		await stream.close();
	}
});
