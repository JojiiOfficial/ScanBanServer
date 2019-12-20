package main

//FilterProcessor processes filter
type FilterProcessor struct {
	ipworker   chan IPDataResult
	builder    *FilterBuilder
	queueCount int
	id         int
}

func (processor *FilterProcessor) start() {
	processor.ipworker = make(chan IPDataResult, 1)
	go (func() {
		for {
			ipdata := <-processor.ipworker
			processor.builder.handleIP(ipdata)
			processor.queueCount--
		}
	})()
}

func (processor *FilterProcessor) addIP(ipdata IPDataResult) {
	processor.queueCount++
	processor.ipworker <- ipdata
}

//FilterHandler handles ips and filter
type FilterHandler struct {
	processors     []*FilterProcessor
	processorCount int
}

func (handler *FilterHandler) init(processorCount int, builder *FilterBuilder) {
	handler.processorCount = processorCount
	for i := 0; i < processorCount; i++ {
		processor := FilterProcessor{
			builder: builder,
			id:      i + 1,
		}
		handler.processors = append(handler.processors, &processor)
		processor.start()
	}
}

func (handler *FilterHandler) addIP(ipData IPDataResult) {
	var mostNonBusyProcessor *FilterProcessor
	//Get processor with smallest queue
	for i, processor := range handler.processors {
		if i == 0 {
			mostNonBusyProcessor = handler.processors[0]
		} else {
			if processor.queueCount < mostNonBusyProcessor.queueCount {
				mostNonBusyProcessor = handler.processors[i]
			}
		}
	}
	mostNonBusyProcessor.addIP(ipData)
}
