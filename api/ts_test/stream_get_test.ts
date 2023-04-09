import { test } from 'node:test';
// import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('stream get success', async () => {
	const tc = new TestClient();

	const create = await tc.client.createTestType({text: "foo"});

	const stream = await tc.client.streamGetTestType(create.id);

	console.log(await stream.read());

	await tc.client.updateTestType(create.id, {text: "bar"});

	console.log(await stream.read());

	await stream.abort();
});
