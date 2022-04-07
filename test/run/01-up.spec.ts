import { api } from "./src/api";

describe('ping', () => {
      test("earth rest api ping", async () => {
        const resp = await api.get('/api/v1/ping');
        expect(resp.status).toEqual(200);
      });
  });
  