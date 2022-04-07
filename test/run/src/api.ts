import { URLSearchParams } from 'url';

const fetch = require('node-fetch');

// const root_domain = process.env.EARTH_REST_API_TEST_ROOT_DOMAIN || 'earth-rest-api.fi';
const root_domain = 'localhost' || 'earth-rest-api.fi';

export interface APIResponse {
  details?: any;
  error?: string;
  status_code?: number;
}

const headersPostJSON = {
  'content-type': 'application/json',
};

async function jsonResponse<T>(resp: Response): Promise<any> {
  if (resp.ok) {
    return resp.json();
  }
  if ((resp.headers.get('content-type') || '').startsWith('application/json')) {
    const data = await resp.json();
    throw data;
  }
  throw `${resp.status} ${resp.statusText} - ${resp.url}`;
}

class API {
  fetch(url: string, options?: RequestInit): Promise<Response> {
    return fetch(url, options);
  }

  url(path: string, query?: any): string {
    let url = 'http://' + root_domain + ':8080';
    if (path && path[0] !== '/') {
      path = '/' + path;
    }
    url += path || '/';
    if (query) {
      url += '?' + new URLSearchParams(query).toString();
    }
    return url;
  }

  get(path: string, query?: any): Promise<Response> {
    return this.fetch(this.url(path, query));
  }

  json<T = APIResponse>(path: string, query?: any): Promise<T> {
    return this.get(path, query).then(jsonResponse);
  }

  post<T = APIResponse>(path: string, body?: any): Promise<T> {
    return this.postQuery(path, undefined, body);
  }

  postQuery<T = APIResponse>(path: string, query?: any, body?: any): Promise<T> {
    return this.fetch(this.url(path, query), {
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
      headers: body ? headersPostJSON : undefined,
    }).then(jsonResponse);
  }

  update<T = APIResponse>(path: string, body?: any): Promise<T> {
    return this.updateQuery(path, undefined, body);
  }

  updateQuery<T = APIResponse>(path: string, query?: any, body?: any): Promise<T> {
    return this.fetch(this.url(path, query), {
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
      headers: body ? headersPostJSON : undefined,
    }).then(jsonResponse);
  }

  delete(path: string, query?: any): Promise<Response> {
    return this.fetch(this.url(path, query), {
      method: 'DELETE',
    }).then(jsonResponse);
  }
}

export const api = new API();
