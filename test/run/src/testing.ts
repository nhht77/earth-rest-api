import { api, APIResponse } from './api';

import moment = require('moment');

import { format as pretty_format } from 'pretty-format';
import { City, Continent, Country, UserMinimal } from '../types/types';

export const pretty_print = (val: unknown) => {
  console.log(pretty_format(val));
};

export class Testing {
  state = new Map<string, any>();

  constructor() {}

  expectDetail(expectfn: () => void, error_message: string) {
    try {
      expectfn();
    } catch (err) {
      throw new Error(err + '\n\nDetail: ' + error_message);
    }
  }

  matchArrayProperty<T>(id_prop: keyof T, a: T[], b: T[]) {
    expect(a.length).toEqual(b.length);
    a.forEach((a_child) => {
      let b_child = b.find((iter) => iter[id_prop] === a_child[id_prop]);
      if (!b_child) {
        throw new Error(`failed to find ${id_prop} ${a_child[id_prop]} child from other array`);
      }
      this.match(a_child, b_child);
    });
  }

  matchArray<T>(a: T[], b: T[]) {
    expect(a.length).toEqual(b.length);
    a.forEach((a_child, index) => {
      this.match(a_child, b[index]);
    });
  }

  match<T>(a: T, b: T) {
    Object.keys(a).forEach((prop) => {
      if (typeof a[prop] === 'object' && a[prop] !== null) {
        this.match(a[prop], b[prop]);
      } else {
        if (b === undefined) {
          throw `b is undefined for property '${prop}'`;
        }
        const value_a = a[prop];
        const value_b = b[prop];
        this.expectDetail(() => expect(value_a).toEqual(value_b), `match failed for property '${prop}'`);
      }
    });
  }

  responseOK(resp: APIResponse | Response) {
    // @ts-ignore success responses should not have these properties
    if (resp.status_code === undefined && resp.status === undefined && resp.error === undefined) {
      return;
    }

    // @ts-ignore
    if (typeof resp.status_code === 'number') {
      resp = resp as APIResponse;
      if (resp.error) {
        pretty_print(resp);
        expect(resp.error).toEqual('');
      }
      expect(resp.status_code).toEqual(200);
    } else {
      resp = resp as Response;
      expect(resp.status).toEqual(200);
    }
  }

  stateAppend<T>(key: string, obj: T) {
    if (!this.state.has(key)) {
      this.state.set(key, []);
    }
    let arr = this.state.get(key) as T[];
    arr.push(obj);
  }

  removeProperties<T>(obj: T, ...prop: (keyof T)[]): T {
    let out = Object.assign({}, obj) as T;
    prop.forEach((key) => delete out[key]);
    return out;
  }

  removePropertiesArray<T>(obj: T[], ...prop: (keyof T)[]): T[] {
    return obj.map((iter) => this.removeProperties(iter, ...prop));
  }
}

export const testing = new Testing();

export class EarthAPITesting extends Testing {
  constructor() {
    super();
  }

  creator: UserMinimal = {
    email: 'makkara.sinappi@gmail.com',
    name: 'Makkara Sinappi',
  }

  data = {
    Continent: {
      name: 'Europe',
      type: 3,
      area_by_km2: 366033131,
      creator: this.creator,
    } as Continent,
    Country: {
      name: 'Finland',
      details: {
        phone_code: '358',
        iso_code: 'FI / FIN',
        currency: '€',
        continent: {},
      },
      creator: this.creator,
    } as Country,
    City: {
      name: 'Helsinki',
      details: {
        is_capital: true,
      },
      creator: this.creator,
    } as City,
  };

  initializeFromAPI(...t: ('continent' | 'country' | 'city')[]): Promise<any> {
    return Promise.all(
      t.map((obj_type): Promise<any> => {
        switch (obj_type) {
          case 'continent': {
            return api.json<Continent[]>('/api/v1/continents').then((results) => {
              this.data.Continent = results.find((iter) => (iter.uuid = this.data.Continent.uuid));
            });
          }
          case 'country': {
            return api.json<Country[]>('/api/v1/countries').then((results) => {
              this.data.Continent = results.find((iter) => (iter.uuid = this.data.Continent.uuid));
            });
          }
          case 'city': {
            return api.json<Country[]>('/api/v1/cities').then((results) => {
              this.data.Continent = results.find((iter) => (iter.uuid = this.data.Continent.uuid));
            });
          }
        }
      })
    );
  }
}

export const earth_testing = new EarthAPITesting();

