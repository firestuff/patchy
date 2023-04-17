import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream list success', async () => {
	const tc = new TestClient();

	await tc.client.createTestType({text: "foo"});

	const stream = await tc.client.streamListTestType({sorts: ["+text"]});

	try {
		const s1 = await stream.read();
		assert(s1);
		assert.equal(s1.length, 1);
		assert.equal(s1[0]!.text, "foo");

		const create2 = await tc.client.createTestType({text: "bar"});

		const s2 = await stream.read();
		assert(s2);
		assert.equal(s2.length, 2);
		assert.equal(s2[0]!.text, "bar");
		assert.equal(s2[1]!.text, "foo");

		await tc.client.updateTestType(create2.id, {text: "zig"});

		const s3 = await stream.read()!;
		assert(s3);
		assert.equal(s3.length, 2);
		assert.equal(s3[0]!.text, "foo");
		assert.equal(s3[1]!.text, "zig");
	} finally {
		await stream.close();
	}
});
