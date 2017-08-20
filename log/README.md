Simple Log library supporting features:

 - namespaces helping recognize where logged message was invoked, 
 - different types of logs such as:
    - **Info**, when you want express to user of your code that something 
      important (not negative) has happened. It should be used for informing
      about rare situation in your code, such as database connection has ben 
      established.
      
    - **Debug**, it meant to be verbose, like every few lines of code, when 
      you decide that part of your code implementation did something 
      important for internal state of your library/code.
      
    - **Error**, when your implementation receive error which is handled by
      your code, but additionally you would like to store information about
      that particular error.    