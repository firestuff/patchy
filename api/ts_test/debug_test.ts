import { strict as assert } from 'node:assert';
import { TestClient } from './util.js';

const tc = new TestClient();
const dbg = await tc.client.debugInfo();
assert(dbg);
