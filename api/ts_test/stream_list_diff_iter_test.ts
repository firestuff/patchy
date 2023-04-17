import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream list diff iter success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: 'foo'});

	const stream = await tc.client.streamListTestType({stream: 'diff'});

	try {
		await tc.client.updateTestType(create.id, {text: 'bar'});

		const objs = [];

		for await (const list of stream) {
			assert.equal(list.length, 1);

			objs.push(list[0]);

			if (objs.length == 2) {
				await stream.abort();
			}
		}

		assert.deepEqual(objs.map(x => x!.text), ['foo', 'bar']);
	} finally {
		await stream.close();
	}
});