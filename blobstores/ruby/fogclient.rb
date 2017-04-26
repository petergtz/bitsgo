#!/usr/bin/env ruby

require 'bundler'
Bundler.require


options =  JSON.parse(ARGV[0], :symbolize_names => true)
STDERR.puts options
connection = Fog::Storage.new(options)
command = ARGV[1]
STDERR.puts "command: #{command}"

directory_key = ARGV[2]
path = ARGV[3]

case command
when "Exists"
    # connection.directories.get(directory_key, 'limit' => 1, max_keys: 1)
    # connection.directories.create(key: @directory_key, public: false)
  puts !connection.directories.get(directory_key, 'limit' => 1, max_keys: 1).files.head(path).nil?
else
  puts 'Unexpected command'
end
