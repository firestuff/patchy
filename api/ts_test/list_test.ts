import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('list prev success', async () => {
	const tc = new TestClient();

	await tc.client.createTestType({text: "foo"});
	await tc.client.createTestType({text: "bar"});

	const list1 = await tc.client.listTestType();
	assert.deepStrictEqual(list1.map(x => x.text).sort(), ["bar", "foo"]);

	// TODO: Mutate list1 to be sneaky

	const list2 = await tc.client.listTestType({prev: list1});
	assert.deepStrictEqual(list2.map(x => x.text).sort(), ["bar", "foo"]);
});
