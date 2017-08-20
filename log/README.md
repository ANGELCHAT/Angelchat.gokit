# log

Package log features:
 - message tags, helps recognize where logged message was invoked, 
 - three levels of logging: Info, Debug, Error
   
          
 ## func Info
 ``` go
 func Info(tag, msg string, args ...interface{}) 
 ```
 Use it when you want emphasize that something important (not negative) 
 has happened. It should be used for informing about rare situation in 
 your code, such as database connection has ben established.
 
 ## func Debug
 ``` go
 func Debug(tag, msg string, args ...interface{}) 
 ```
 it meant to be verbose, like every few lines of code, when 
 you decide that part of your code implementation did something 
 important for internal state of your library/code.
 
 ## func Error
 ``` go
 func Error(tag string, e error) 
 ```
 when your implementation receive error which is handled by
 your code, but additionally you would like to store information about
 that particular error.