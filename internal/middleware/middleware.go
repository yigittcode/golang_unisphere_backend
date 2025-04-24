package middleware

// Middleware fonksiyonları burada olacak
// Örnek:
// func LoggerMiddleware() gin.HandlerFunc {
// 	return func(c *gin.Context) {
// 		// İstek geldiğinde
// 		start := time.Now()
//
// 		// Sonraki middleware'e geç
// 		c.Next()
//
// 		// İstek tamamlandığında
// 		duration := time.Since(start)
// 		log.Printf("%s %s %s %d %s", c.Request.Method, c.Request.URL.Path, c.ClientIP(), c.Writer.Status(), duration)
// 	}
// }
