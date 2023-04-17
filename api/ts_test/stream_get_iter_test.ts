import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream get iter success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: 'foo'});

	const stream = await tc.client.streamGetTestType(create.id);

	try {
		await tc.client.updateTestType(create.id, {text: 'bar'});

		const objs = [];

		for await (const obj of stream) {
			objs.push(obj);

			if (objs.length == 2) {
				await stream.abort();
			}
		}

		assert.equal(objs.length, 2);
		assert.equal(objs[0]!.text, 'foo');
		assert.equal(objs[1]!.text, 'bar');
	} finally {
		await stream.close();
	}
});
