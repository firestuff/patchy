import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('find success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: 'foo'});

	const find = await tc.client.findTestType(create.id.substring(0, 4));
	assert.equal(find.text, 'foo');
});
