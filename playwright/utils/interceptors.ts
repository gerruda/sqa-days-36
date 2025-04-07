import { Page, Request } from '@playwright/test';
import { Schema } from 'ajv';

type RUMRequest = {
  url: string;
  method: string;
  payload: any;
  status?: number;
  error?: string;
};

export class RUMInterceptor {
 requests: RUMRequest[] = [];
  errors: Error[] = [];
  readonly rumEndpoint: string;

  constructor(page: Page, endpointPattern: string) {
    this.rumEndpoint = endpointPattern;
    this._setupInterception(page);
  }

  async _setupInterception(page: Page) {
    await page.route(this.rumEndpoint, async (route, request) => {
      try {
        const postData = request.postData();
        const payload = postData ? JSON.parse(postData) : null;

        this.requests.push({
          url: request.url(),
          method: request.method(),
          payload,
          status: request.response()?.status()
        });

        await route.continue();
      } catch (error) {
        this.errors.push(error as Error);
        await route.abort();
      }
    });
  }

  getRequests(): RUMRequest[] {
    return this.requests;
  }

  getErrors(): Error[] {
    return this.errors;
  }

  clear() {
    this.requests = [];
    this.errors = [];
  }

  validateSchema(schema: Schema): boolean[] {
    const ajv = new Ajv(); // Не забудьте установить пакет ajv
    return this.requests.map(req =>
      ajv.validate(schema, req.payload)
    );
  }
}