package tests

import (
"io"
"net/http"
"testing"
"github.com/ozertech/ozon-api-client" // Условный импорт
"github.com/zdz1715/allure-go"
)

const (
	baseURL       = "https://www.ozon.ru"
	apiServiceURL = "https://api.internal.ozon.ru/v1/data"
	dbCheckURL    = "https://dbauth.ozon.ru/verify"
)

func (s TestSuite) TestOzonRum(t provider.T)  {
			t.Epic("Ozon Data Flow"),
			t.Feature("Data Consistency Check"),
			t.Description("Проверка сквозного потока данных через сервисы Ozon"),
			t.WithNewStep("🔗 Отправляем запрос на создание аннотации в сервис", func(sCtx provider.StepCtx) {
				ctx := context_provider.WithProvider(s.ctx, sCtx)
				defer ctx.Done()

				var targetHeader string
				var dataID string

				// Шаг 1: Отправка запроса в Ozon.ru и получение заголовка
				t.WithNewStep("Запрос к Ozon.ru", func(sCtx provider.StepCtx) {
					req, _ := http.NewRequest("GET", baseURL+"/some-product-page", nil)
					client := &http.Client{}
					resp, err := client.Do(req)
					sCtx.Require().NoError(err,"Ошибка при выполнении запроса: %v", err)

					defer resp.Body.Close()

					sCtx.Require().Equal(http.StatusOK, resp.StatusCode, "Неверный статус код: %d", resp.StatusCode)

					targetHeader = resp.Header.Get("Data-Id")
					sCtx.Require().NotEmpty(targetHeader, "Заголовок Data-Id")
				})

				// Шаг 2: Получение данных из внутреннего сервиса
				var serviceData api_pb.GetDataByIDResponce
				t.WithNewStep("Запрос данных из сервиса", func(sCtx provider.StepCtx) {
					req := api_pb.GetDataByIDRequest{
						DataID: dataID
					}
					headers := map[string]string{"Authorization": "Bearer "+ozonapi.GetAPIToken()}

					resp, err := ozonapiService.GetDataByID(s.ctx, req, headers)
						sCtx.Require().NoError(err,"Ошибка сервиса получения метрик: %v", err)
						sCtx.Require().NotNil(resp, "Данные не найдены для ID: %s", dataID)
				})

				reqRum := rumService.rumRequest{
					ID: dataID,
					Metrics: []rumService.rumMetrics{
						Lcp: {
							value: random.Int32(),
						},
					},
				}

				// Шаг 3: Отправка в сервис хранения RUM метрик
				t.WithNewStep("Валидация в БД", func(sCtx provider.StepCtx) {
				resp, err := rumService(s.ctx, reqRum)
					sCtx.Require().NoError(err,"Ошибка сервиса отправки метрик: %v", err)
					sCtx.Require().NotNil(resp, "Ответ не пустой")
				})

				// Шаг 4: Проверка записи в БД
				t.WithNewStep("Валидация с данными из БД метрик", func(sCtx provider.StepCtx) {
					bdData, err := rumService.getByID(s.ctx, dataID)
					sCtx.Require().NoError(err,"Ошибка сервиса получения данных из БД: %v", err)
					sCtx.Require().NotNil(bdData, "Ответ не пустой")

					sCtx.Assert().Equal(reqRum.Metrics[0].Lcp, bdData.Lcp, "Поле LCP")
				})
			}),
		)
	}
}