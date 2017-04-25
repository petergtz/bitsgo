#!/usr/bin/env ruby

require 'bundler'
Bundler.require

puts "Hello World"

options =  ARGV[0]
puts options
connection = Fog::Storage.new(options)
command = ARGV[1]

path =  ARGV[2]

case command
when "Exists"
    # connection.directories.get(directory_key, 'limit' => 1, max_keys: 1)
    # connection.directories.create(key: @directory_key, public: false)
  connection.directories.get(directory_key, 'limit' => 1, max_keys: 1).files.head(key).nil?
else
  puts 'Unexpected command'
end
