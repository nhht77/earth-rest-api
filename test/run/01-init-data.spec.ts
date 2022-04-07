import { api } from "./src/api";
import { earth_testing } from "./src/testing";
import { City, CityDetails, Continent, Country, CountryDetails } from "./types/earth";

describe('ping', () => {
  test("earth rest api ping", async () => {
    const resp = await api.get('/api/v1/ping');
    expect(resp.status).toEqual(200);
  });
});

describe('earth rest api / continent', () => {
  test(`create ${earth_testing.data.Continent.name}`, async () => {
    const continent = await api.post<Continent>('/api/v1/continent/create', earth_testing.data.Continent);
    expect(continent.uuid).toBeDefined();
    earth_testing.match(
      earth_testing.removeProperties(earth_testing.data.Continent, 'updated'),
      earth_testing.removeProperties(continent, 'updated'),
    );
    earth_testing.data.Continent = continent;
  });
  
  test(`update ${earth_testing.data.Continent.name}`, async () => {
    const updated: { [k in keyof Continent]: any } = Object.assign(
      {},
      earth_testing.data.Continent,
      {
        name:"Europe #Updated",
        area_by_km2:266033131,
      },
    );

    
    const continent = await api.update<Continent>('/api/v1/continent/update', updated);
    expect(continent.uuid).toEqual(earth_testing.data.Continent.uuid);

    earth_testing.match(
      earth_testing.removeProperties(updated, 'updated'),
      earth_testing.removeProperties(continent, 'updated'),
    );
    earth_testing.data.Continent = continent;
  });
  test('create for delete', async () => {
    let src = earth_testing.removeProperties(
      earth_testing.data.Continent,
      'uuid',
      'created',
      'updated',
    );
    src.type = 1
    src.name = "Asia"
    src.area_by_km2 = 266033131
    const continent = await api.post<Continent>('/api/v1/continent/create', src);
    expect(continent.uuid).toBeDefined();
    earth_testing.match(src, continent);
    earth_testing.stateAppend('delete_continent', continent);
  });
  test('delete', async () => {
    let delete_targets = earth_testing.state.get('delete_continent') as Continent[];
    let requests = delete_targets.map(async (delete_target) => {
      return api.delete(`/api/v1/continent/delete?uuid=${delete_target.uuid}`);
    });
    (await Promise.all(requests)).forEach((resp) => {
      expect(resp.body).toBeUndefined();
    });
    earth_testing.state.delete('delete_continent');
  });
  test('list', async () => {
    const results = await api.json<Continent[]>('/api/v1/continents');
    expect(results.length).toEqual(1);
  });
});

describe('earth rest api / country', () => {
  test(`create ${earth_testing.data.Country.name}`, async () => {

    const src: Country = {
      continent_uuid: earth_testing.data.Continent.uuid,
      ...earth_testing.data.Country,
    }
    
    const country = await api.post<Country>('/api/v1/country/create', src);
    expect(country.uuid).toBeDefined();
    earth_testing.match(
      earth_testing.removeProperties(src, 'updated'),
      earth_testing.removeProperties(country, 'updated'),
    );
    earth_testing.data.Country = country;
  });
  
  test(`update ${earth_testing.data.Country.name}`, async () => {
    const updated: { [k in keyof Country]: any } = Object.assign(
      {},
      earth_testing.data.Country,
      { name:"Finland #Updated"} as Country,
    );

    
    const country = await api.update<Country>('/api/v1/country/update', updated);
    expect(country.uuid).toEqual(earth_testing.data.Country.uuid);

    earth_testing.match(
      earth_testing.removeProperties(updated, 'updated'),
      earth_testing.removeProperties(country, 'updated'),
    );
    earth_testing.data.Country = country;
  });

  test('create for delete', async () => {
    let src = earth_testing.removeProperties(
      earth_testing.data.Country,
      'uuid',
      'created',
      'updated',
    );
    src.name = "French",
    src.details = {
      phone_code: "68",
      iso_code: "Fr / FR",
      currency: "€",
    } as CountryDetails

    const country = await api.post<Country>('/api/v1/country/create', src);
    expect(country.uuid).toBeDefined();
    earth_testing.match(src, country);
    earth_testing.stateAppend('delete_country', country);
  });

  test('delete', async () => {
    let delete_targets = earth_testing.state.get('delete_country') as Country[];
    let requests = delete_targets.map(async (delete_target) => {
      return api.delete(`/api/v1/country/delete?uuid=${delete_target.uuid}`);
    });
    (await Promise.all(requests)).forEach((resp) => {
      expect(resp.body).toBeUndefined();
    });
    earth_testing.state.delete('delete_country');
  });
  test('list', async () => {
    const results = await api.json<Continent[]>('/api/v1/countries');
    expect(results.length).toEqual(1);
  });
});

describe('earth rest api / city', () => {
  test(`create ${earth_testing.data.City.name}`, async () => {

    const src: City = {
      continent_uuid: earth_testing.data.Continent.uuid,
      country_uuid: earth_testing.data.Country.uuid,
      ...earth_testing.data.City,
    }
    
    const city = await api.post<City>('/api/v1/city/create', src);
    expect(city.uuid).toBeDefined();
    earth_testing.match(
      earth_testing.removeProperties(src, 'updated'),
      earth_testing.removeProperties(city, 'updated'),
    );
    earth_testing.data.City = city;
  });
  
  test(`update ${earth_testing.data.City.name}`, async () => {
    const updated: { [k in keyof City]: any } = Object.assign(
      {},
      earth_testing.data.City,
      { name:"Helsinki #Updated"} as City,
    );

    
    const city = await api.update<City>('/api/v1/city/update', updated);
    expect(city.uuid).toEqual(earth_testing.data.City.uuid);

    earth_testing.match(
      earth_testing.removeProperties(updated, 'updated'),
      earth_testing.removeProperties(city, 'updated'),
    );
    earth_testing.data.City = city;
  });

  test('create for delete', async () => {
    let src = earth_testing.removeProperties(
      earth_testing.data.City,
      'uuid',
      'created',
      'updated',
    );
    src.name = "Seinäjoki",
    src.details = { is_capital: false} as CityDetails

    const city = await api.post<City>('/api/v1/city/create', src);
    expect(city.uuid).toBeDefined();
    earth_testing.match(src, city);
    earth_testing.stateAppend('delete_city', city);
  });

  test('delete', async () => {
    let delete_targets = earth_testing.state.get('delete_city') as City[];
    let requests = delete_targets.map(async (delete_target) => {
      return api.delete(`/api/v1/city/delete?uuid=${delete_target.uuid}`);
    });
    (await Promise.all(requests)).forEach((resp) => {
      expect(resp.body).toBeUndefined();
    });
    earth_testing.state.delete('delete_city');
  });
  test('list', async () => {
    const results = await api.json<Continent[]>('/api/v1/cities');
    expect(results.length).toEqual(1);
  });
});