
// General class

export const typed_assign = <T>(target: T, options: Partial<T>) => {
    if (target && options) {
        Object.assign(target, options);
    }
};


export class UserMinimal {
    email: string;
    name: string;

    constructor(options?: Partial<UserMinimal>) {
        typed_assign(this, options);
    }
}

export enum ContinentType {
    ContinentTypeInvalid  = 0,
	ContinentTypeAsia  = 1,
	ContinentTypeAfrica  = 2,
	ContinentTypeEurope  = 3,
	ContinentTypeNorth_America  = 4,
	ContinentTypeSouth_America  = 5,
	ContinentTypeOceania  = 6,
	ContinentTypeAntarctica  = 7,
}

export class Continent  {
    uuid?: string;
    name?: string;
    type?: ContinentType;
    area_by_km2?: number;
    creator?: UserMinimal;

    created?: string;
    updated?: string;

    constructor(options?: Partial<Continent>) {
        typed_assign(this, options);
    }
}

export class CountryDetails {
    phone_code?: string;
    iso_code?: string;
    currency?: string;
    continent: Continent;

    constructor(options?: Partial<CountryDetails>) {
        typed_assign(this, options);
    }
}

export class Country {
    continent_uuid?:string;
    uuid?:string;
    name?:string;
    details?:CountryDetails;
    created?:string;
    updated?:string;
    creator?:UserMinimal;

    constructor(options?: Partial<Country>) {
        typed_assign(this, options);
    }
}

export class CityDetails {
    is_capital?:boolean;
    continent?:Continent;
    country?:Country;

    constructor(options?: Partial<CityDetails>) {
        typed_assign(this, options);
    }
}

export class City {
    continent_uuid?: string;
    country_uuid?: string;
    uuid?: string;

    name?: string;
    details?: CityDetails;

    created?: string;
    updated?: string;

    creator?: UserMinimal;

    constructor(options?: Partial<City>) {
        typed_assign(this, options);
    }
}
