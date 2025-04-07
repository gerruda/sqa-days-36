import { test, expect } from '@playwright/test'
import { RUMInterceptor } from './utils/interceptors';

test.describe('RUM Tests', () => {
  let rumInterceptor: RUMInterceptor;

  test.beforeEach(async ({ page }) => {
    rumInterceptor = new RUMInterceptor(page, '**/api/rum-metrics');
    await page.goto('https://ozon.ru');
  });

  test('Тело запроса RUM', async ({ page }) => {
    const requests = rumInterceptor.getRequests();

    // Проверка количества запросов
    expect(requests.length, 'Есть запросы RUM').toBeGreaterThan(0);
    expect.soft(rumInterceptor.getErrors(), 'Нет ошибок API').toHaveLength(0);

    // Проверка JSON-схемы
    const validationResults = rumInterceptor.validateSchema(rumSchema);
    expect(validationResults.every(Boolean), 'Запросы соответствуют JSONсхеме').toBeTruthy();

    // Проверка конкретных значений
    // Выполняем JavaScript в консоли браузера
    const heapSize: number = await page.evaluate(() => {
      // Используем тип any для обхода TS-ошибок (Chrome-specific API)
      return (performance as any).memory?.usedJSHeapSize;
    });

    // Проверяем, что значение получено
    expect(heapSize).toBeDefined();

    requests.forEach(req => {
      expect.soft(req.payload, 'есть поле eventType').toHaveProperty('eventType');
      expect.soft(req.payload, 'есть поле usedJSHeapSize').toHaveProperty('usedJSHeapSize');
      expect.soft(req.payload.usedJSHeapSize, 'поле usedJSHeapSize корректно').toBeLessThanOrEqual(heapSize);
    });
  });
});