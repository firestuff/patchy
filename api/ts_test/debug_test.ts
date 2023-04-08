import { test } from 'node:test';
import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

test('debug fetch success', async () => {
	const tc = new TestClient();
	const dbg = await tc.client.debugInfo();
	assert(dbg);
});
