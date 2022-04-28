from concurrent import futures
import grpc
import stockapi_pb2
import stockapi_pb2_grpc
import time

class Listener(stockapi_pb2_grpc.StockPredictionServicer):
    def __init__(self, *args, **kwargs):
        pass
    
    def getStock(self, request, context):
        stockName = request.name
        print(stockName)
        date = request.date
        print(date)
        data = []
        for i in range(50):
            data.append(float(i))
        result = {'data': data, 'prediction': 20, 'recomandation': "buy", 'status': ""}
        return stockapi_pb2.APIReturn(**result)

def serve():
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))
    stockapi_pb2_grpc.add_StockPredictionServicer_to_server(Listener(), server)
    server.add_insecure_port('[::]:50002')
    server.start()
    server.wait_for_termination()

if __name__ == '__main__':
    serve()
        