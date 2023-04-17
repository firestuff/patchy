import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('list opts success', async () => {
	const tc = new TestClient();

	await tc.client.createTestType({text: 'foo'});
	await tc.client.createTestType({text: 'bar'});
	await tc.client.createTestType({text: 'zig'});
	await tc.client.createTestType({text: 'aaa'});

	const list = await tc.client.listTestType({
		limit: 1,
		offset: 1,
		sorts: ['+text'],
		filters: [
			{
				path: 'text',
				op: 'gt',
				value: 'aaa',
			},
		],
	});
	assert.deepStrictEqual(list.map(x => x.text), ['foo']);
});
