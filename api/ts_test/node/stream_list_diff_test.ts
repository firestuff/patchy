import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream list diff success', async () => {
	const tc = new TestClient();

	await tc.client.createTestType({text: 'foo'});
	await tc.client.createTestType({text: 'zig'});
	await tc.client.createTestType({text: 'aaa'});

	const stream = await tc.client.streamListTestType({stream: 'diff', sorts: ['+text']});

	try {
		const s1 = await stream.read();
		assert(s1);
		assert.deepEqual(s1.map(x => x.text), ['aaa', 'foo', 'zig']);

		const create2 = await tc.client.createTestType({text: 'bar'});

		const s2 = await stream.read();
		assert(s2);
		assert.deepEqual(s2.map(x => x.text), ['aaa', 'bar', 'foo', 'zig']);

		await tc.client.updateTestType(create2.id, {text: 'zag'});

		const s3 = await stream.read()!;
		assert(s3);
		assert.deepEqual(s3.map(x => x.text), ['aaa', 'foo', 'zag', 'zig']);
	} finally {
		await stream.close();
	}
});
